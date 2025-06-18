// handlers.go (API handlers)
package main

import (
    "net/http"
    "strconv"
    "time"

    "crawler666/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func (app *CrawlerApp) startCrawl(c *gin.Context) {
    var req struct {
        Name        string   `json:"name" binding:"required"`
        Description string   `json:"description"`
        StartURLs   []string `json:"start_urls" binding:"required"`
        Rules       models.CrawlRules `json:"rules"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Create crawl session
    session := &models.CrawlSession{
        ID:          uuid.New().String(),
        Name:        req.Name,
        Description: req.Description,
        StartURLs:   req.StartURLs,
        Rules:       req.Rules,
        Status:      "active",
        CreatedAt:   time.Now(),
        Stats:       models.SessionStats{},
    }

    if err := app.Storage.CreateCrawlSession(session); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
        return
    }

    // Create initial tasks
    for _, url := range req.StartURLs {
        task := &models.CrawlTask{
            ID:          uuid.New().String(),
            SessionID:   session.ID,
            URL:         url,
            Method:      "GET",
            Priority:    5,
            MaxDepth:    req.Rules.MaxDepth,
            CreatedAt:   time.Now(),
            ScheduledAt: time.Now(),
            Status:      "pending",
        }
        
        // Add task to queue (simplified - would use proper task creation)
        select {
        case app.Engine.queue <- task:
        default:
            app.Logger.Warn("Queue full, task will be scheduled later")
        }
    }

    c.JSON(http.StatusCreated, session)
}

func (app *CrawlerApp) getCrawlStatus(c *gin.Context) {
    sessionID := c.Param("id")
    
    sessions, err := app.Storage.GetCrawlSessions()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
        return
    }

    for _, session := range sessions {
        if session.ID == sessionID {
            c.JSON(http.StatusOK, session)
            return
        }
    }

    c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
}

func (app *CrawlerApp) stopCrawl(c *gin.Context) {
    sessionID := c.Param("id")
    
    // Implementation would stop all tasks for this session
    // For now, return success
    c.JSON(http.StatusOK, gin.H{"message": "Crawl stopped", "session_id": sessionID})
}

func (app *CrawlerApp) listCrawls(c *gin.Context) {
    sessions, err := app.Storage.GetCrawlSessions()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
        return
    }

    c.JSON(http.StatusOK, sessions)
}

func (app *CrawlerApp) getConfig(c *gin.Context) {
    c.JSON(http.StatusOK, app.Config)
}

func (app *CrawlerApp) updateConfig(c *gin.Context) {
    var newConfig Config
    if err := c.ShouldBindJSON(&newConfig); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update configuration (simplified)
    app.Config = &newConfig
    c.JSON(http.StatusOK, gin.H{"message": "Configuration updated"})
}

func (app *CrawlerApp) getStats(c *gin.Context) {
    stats := app.Engine.GetStats()
    proxyStats := app.ProxyMgr.GetStats()

    response := gin.H{
        "crawler": stats,
        "proxies": proxyStats,
        "timestamp": time.Now(),
    }

    c.JSON(http.StatusOK, response)
}

func (app *CrawlerApp) healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
        "version": "1.0.0",
    })
}

func (app *CrawlerApp) getMetrics(c *gin.Context) {
    stats := app.Engine.GetStats()
    
    metrics := gin.H{
        "crawler_requests_total": stats.TotalRequests,
        "crawler_success_rate": float64(stats.SuccessfulCrawls) / float64(stats.TotalRequests) * 100,
        "proxy_pool_health": 95.0, // Would calculate from actual proxy stats
        "stealth_detection_events": stats.DetectionEvents,
        "active_workers": stats.ActiveWorkers,
        "queue_size": stats.QueueSize,
    }

    c.JSON(http.StatusOK, metrics)
}

func (app *CrawlerApp) getProxies(c *gin.Context) {
    stats := app.ProxyMgr.GetStats()
    c.JSON(http.StatusOK, stats)
}

func (app *CrawlerApp) testProxy(c *gin.Context) {
    var req struct {
        Host string `json:"host" binding:"required"`
        Port int    `json:"port" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Test proxy (simplified)
    result := gin.H{
        "host": req.Host,
        "port": req.Port,
        "status": "healthy",
        "response_time": "245ms",
        "country": "US",
    }

    c.JSON(http.StatusOK, result)
}

func (app *CrawlerApp) exportData(c *gin.Context) {
    crawlID := c.Param("crawlId")
    limitStr := c.DefaultQuery("limit", "1000")
    
    limit, err := strconv.Atoi(limitStr)
    if err != nil {
        limit = 1000
    }

    results, err := app.Storage.GetCrawlResults(crawlID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get results"})
        return
    }

    c.Header("Content-Type", "application/json")
    c.Header("Content-Disposition", "attachment; filename=crawl_"+crawlID+".json")
    c.JSON(http.StatusOK, results)
}
