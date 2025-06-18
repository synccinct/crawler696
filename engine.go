// engine.go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "crawler666/internal/models"
    "crawler666/pkg/proxy"
    "crawler666/pkg/stealth"
    "crawler666/pkg/storage"

    "github.com/sirupsen/logrus"
)

type CrawlerEngine struct {
    config     *CrawlerConfig
    storage    storage.Interface
    proxyMgr   *proxy.Manager
    stealthEng *stealth.Engine
    logger     *logrus.Logger
    
    workers    map[string]*Worker
    scheduler  *Scheduler
    queue      chan *models.CrawlTask
    results    chan *models.CrawlResult
    
    mu         sync.RWMutex
    running    bool
    stats      *CrawlStats
}

type Worker struct {
    ID       string
    Engine   *CrawlerEngine
    ctx      context.Context
    cancel   context.CancelFunc
    active   bool
}

type Scheduler struct {
    engine    *CrawlerEngine
    domains   map[string]*DomainState
    mu        sync.RWMutex
}

type DomainState struct {
    LastRequest time.Time
    RequestRate int
    Blocked     bool
    ProxyPool   string
}

type CrawlStats struct {
    TotalRequests     int64
    SuccessfulCrawls  int64
    FailedCrawls      int64
    DetectionEvents   int64
    ActiveWorkers     int
    QueueSize         int
    mu                sync.RWMutex
}

func NewCrawlerEngine(config *CrawlerConfig, storage storage.Interface, 
                     proxyMgr *proxy.Manager, stealthEng *stealth.Engine, 
                     logger *logrus.Logger) *CrawlerEngine {
    
    engine := &CrawlerEngine{
        config:     config,
        storage:    storage,
        proxyMgr:   proxyMgr,
        stealthEng: stealthEng,
        logger:     logger,
        workers:    make(map[string]*Worker),
        queue:      make(chan *models.CrawlTask, config.QueueSize),
        results:    make(chan *models.CrawlResult, config.QueueSize),
        stats:      &CrawlStats{},
    }

    engine.scheduler = &Scheduler{
        engine:  engine,
        domains: make(map[string]*DomainState),
    }

    return engine
}

func (e *CrawlerEngine) StartWorkers(ctx context.Context) {
    e.mu.Lock()
    e.running = true
    e.mu.Unlock()

    // Start result processor
    go e.processResults(ctx)

    // Start scheduler
    go e.scheduler.run(ctx)

    // Start workers
    for i := 0; i < e.config.MaxWorkers; i++ {
        workerID := fmt.Sprintf("worker-%d", i)
        worker := e.createWorker(workerID, ctx)
        e.workers[workerID] = worker
        go worker.run()
    }

    e.logger.Infof("Started %d crawler workers", e.config.MaxWorkers)
}

func (e *CrawlerEngine) createWorker(id string, parentCtx context.Context) *Worker {
    ctx, cancel := context.WithCancel(parentCtx)
    return &Worker{
        ID:     id,
        Engine: e,
        ctx:    ctx,
        cancel: cancel,
        active: true,
    }
}

func (w *Worker) run() {
    w.Engine.logger.Infof("Worker %s started", w.ID)
    
    for {
        select {
        case <-w.ctx.Done():
            w.Engine.logger.Infof("Worker %s stopped", w.ID)
            return
        case task := <-w.Engine.queue:
            w.processTask(task)
        }
    }
}

func (w *Worker) processTask(task *models.CrawlTask) {
    w.Engine.stats.mu.Lock()
    w.Engine.stats.TotalRequests++
    w.Engine.stats.mu.Unlock()

    result := &models.CrawlResult{
        TaskID:    task.ID,
        URL:       task.URL,
        WorkerID:  w.ID,
        StartTime: time.Now(),
    }

    // Get proxy
    proxy, err := w.Engine.proxyMgr.GetProxy(task.URL)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to get proxy: %v", err)
        w.Engine.results <- result
        return
    }

    // Get stealth profile
    profile, err := w.Engine.stealthEng.GenerateProfile(task.URL)
    if err != nil {
        result.Error = fmt.Sprintf("Failed to generate stealth profile: %v", err)
        w.Engine.results <- result
        return
    }

    // Perform crawl
    data, err := w.crawlURL(task.URL, proxy, profile)
    if err != nil {
        result.Error = err.Error()
        w.Engine.stats.mu.Lock()
        w.Engine.stats.FailedCrawls++
        w.Engine.stats.mu.Unlock()
    } else {
        result.Data = data
        result.Success = true
        w.Engine.stats.mu.Lock()
        w.Engine.stats.SuccessfulCrawls++
        w.Engine.stats.mu.Unlock()
    }

    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)

    w.Engine.results <- result
}

func (w *Worker) crawlURL(url string, proxy *proxy.Proxy, profile *stealth.Profile) (*models.CrawlData, error) {
    // Implementation will use stealth browser automation
    // This is a simplified version - full implementation would use chromedp/puppeteer
    
    client := w.Engine.stealthEng.CreateHTTPClient(proxy, profile)
    
    resp, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse content
    data := &models.CrawlData{
        URL:        url,
        StatusCode: resp.StatusCode,
        Headers:    make(map[string]string),
        Timestamp:  time.Now(),
    }

    for k, v := range resp.Header {
        if len(v) > 0 {
            data.Headers[k] = v[0]
        }
    }

    // Read body (simplified - should handle content type parsing)
    body := make([]byte, 1024*1024) // 1MB limit
    n, _ := resp.Body.Read(body)
    data.Content = string(body[:n])

    return data, nil
}

func (e *CrawlerEngine) processResults(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case result := <-e.results:
            // Store result
            if err := e.storage.StoreCrawlResult(result); err != nil {
                e.logger.Errorf("Failed to store crawl result: %v", err)
            }

            // Update metrics based on result
            if result.Error != "" {
                e.logger.Warnf("Crawl failed for %s: %s", result.URL, result.Error)
            } else {
                e.logger.Debugf("Successfully crawled %s", result.URL)
            }
        }
    }
}

func (s *Scheduler) run(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.scheduleNextTasks()
        }
    }
}

func (s *Scheduler) scheduleNextTasks() {
    // Get pending tasks from storage
    tasks, err := s.engine.storage.GetPendingTasks(100)
    if err != nil {
        s.engine.logger.Errorf("Failed to get pending tasks: %v", err)
        return
    }

    for _, task := range tasks {
        // Check domain rate limits
        if s.canScheduleTask(task) {
            select {
            case s.engine.queue <- task:
                s.updateDomainState(task.URL)
            default:
                // Queue full, skip for now
                break
            }
        }
    }
}

func (s *Scheduler) canScheduleTask(task *models.CrawlTask) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()

    domain := extractDomain(task.URL)
    state, exists := s.domains[domain]
    
    if !exists {
        return true
    }

    if state.Blocked {
        return false
    }

    // Check rate limiting
    if time.Since(state.LastRequest) < time.Duration(s.engine.config.RateLimit)*time.Millisecond {
        return false
    }

    return true
}

func (s *Scheduler) updateDomainState(url string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    domain := extractDomain(url)
    state := s.domains[domain]
    if state == nil {
        state = &DomainState{}
        s.domains[domain] = state
    }

    state.LastRequest = time.Now()
    state.RequestRate++
}

func (e *CrawlerEngine) Stop() {
    e.mu.Lock()
    defer e.mu.Unlock()

    if !e.running {
        return
    }

    e.running = false

    // Stop all workers
    for _, worker := range e.workers {
        worker.cancel()
    }

    e.logger.Info("Crawler engine stopped")
}

func (e *CrawlerEngine) GetStats() *CrawlStats {
    e.stats.mu.RLock()
    defer e.stats.mu.RUnlock()

    // Update active workers count
    e.stats.ActiveWorkers = len(e.workers)
    e.stats.QueueSize = len(e.queue)

    return &CrawlStats{
        TotalRequests:    e.stats.TotalRequests,
        SuccessfulCrawls: e.stats.SuccessfulCrawls,
        FailedCrawls:     e.stats.FailedCrawls,
        DetectionEvents:  e.stats.DetectionEvents,
        ActiveWorkers:    e.stats.ActiveWorkers,
        QueueSize:        e.stats.QueueSize,
    }
}

func extractDomain(url string) string {
    // Simplified domain extraction
    // Full implementation would use net/url package
    return url
}
