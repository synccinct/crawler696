// pkg/proxy/manager.go
package proxy

import (
    "errors"
    "fmt"
    "math/rand"
    "net/http"
    "net/url"
    "sync"
    "time"
)

type Manager struct {
    pools       map[string]*Pool
    healthCheck *HealthChecker
    mu          sync.RWMutex
    config      *Config
}

type Config struct {
    Enabled     bool
    Pools       []PoolConfig
    Rotation    int
    HealthCheck int
}

type PoolConfig struct {
    Name      string
    Type      string
    Providers []string
    Endpoints []string
}

type Pool struct {
    Name      string
    Type      string
    Proxies   []*Proxy
    Current   int
    mu        sync.RWMutex
}

type Proxy struct {
    ID          string
    Host        string
    Port        int
    Username    string
    Password    string
    Type        string
    Country     string
    Provider    string
    Healthy     bool
    LastUsed    time.Time
    FailCount   int
    mu          sync.RWMutex
}

type HealthChecker struct {
    manager  *Manager
    interval time.Duration
    testURL  string
}

func NewManager(config *Config) (*Manager, error) {
    manager := &Manager{
        pools:  make(map[string]*Pool),
        config: config,
    }

    // Initialize proxy pools
    for _, poolConfig := range config.Pools {
        pool, err := manager.createPool(poolConfig)
        if err != nil {
            return nil, fmt.Errorf("failed to create pool %s: %v", poolConfig.Name, err)
        }
        manager.pools[poolConfig.Name] = pool
    }

    // Start health checker
    manager.healthCheck = &HealthChecker{
        manager:  manager,
        interval: time.Duration(config.HealthCheck) * time.Second,
        testURL:  "http://httpbin.org/ip",
    }
    go manager.healthCheck.start()

    return manager, nil
}

func (m *Manager) createPool(config PoolConfig) (*Pool, error) {
    pool := &Pool{
        Name:    config.Name,
        Type:    config.Type,
        Proxies: make([]*Proxy, 0),
    }

    // Load proxies from endpoints
    for i, endpoint := range config.Endpoints {
        proxy := &Proxy{
            ID:       fmt.Sprintf("%s-%d", config.Name, i),
            Host:     "proxy.example.com", // Would parse from endpoint
            Port:     8080,
            Type:     config.Type,
            Provider: config.Name,
            Healthy:  true,
        }
        pool.Proxies = append(pool.Proxies, proxy)
    }

    return pool, nil
}

func (m *Manager) GetProxy(targetURL string) (*Proxy, error) {
    if !m.config.Enabled {
        return nil, nil
    }

    m.mu.RLock()
    defer m.mu.RUnlock()

    // Select best pool for target
    poolName := m.selectOptimalPool(targetURL)
    pool, exists := m.pools[poolName]
    if !exists {
        return nil, errors.New("no proxy pools available")
    }

    // Get healthy proxy from pool
    proxy := pool.getHealthyProxy()
    if proxy == nil {
        return nil, errors.New("no healthy proxies available")
    }

    proxy.mu.Lock()
    proxy.LastUsed = time.Now()
    proxy.mu.Unlock()

    return proxy, nil
}

func (m *Manager) selectOptimalPool(targetURL string) string {
    // Simple selection - in production would consider geographic, load, etc.
    for name := range m.pools {
        return name
    }
    return ""
}

func (p *Pool) getHealthyProxy() *Proxy {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.Proxies) == 0 {
        return nil
    }

    // Round-robin selection of healthy proxies
    attempts := 0
    for attempts < len(p.Proxies) {
        proxy := p.Proxies[p.Current]
        p.Current = (p.Current + 1) % len(p.Proxies)
        
        if proxy.Healthy && proxy.FailCount < 5 {
            return proxy
        }
        attempts++
    }

    return nil
}

func (h *HealthChecker) start() {
    ticker := time.NewTicker(h.interval)
    defer ticker.Stop()

    for range ticker.C {
        h.checkAllProxies()
    }
}

func (h *HealthChecker) checkAllProxies() {
    for _, pool := range h.manager.pools {
        for _, proxy := range pool.Proxies {
            go h.checkProxy(proxy)
        }
    }
}

func (h *HealthChecker) checkProxy(proxy *Proxy) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    // Configure proxy for health check
    if proxy.Host != "" {
        proxyURL, err := url.Parse(fmt.Sprintf("http://%s:%d", proxy.Host, proxy.Port))
        if err == nil {
            client.Transport = &http.Transport{
                Proxy: http.ProxyURL(proxyURL),
            }
        }
    }

    resp, err := client.Get(h.testURL)
    
    proxy.mu.Lock()
    if err != nil {
        proxy.Healthy = false
        proxy.FailCount++
    } else {
        proxy.Healthy = true
        proxy.FailCount = 0
        resp.Body.Close()
    }
    proxy.mu.Unlock()
}

func (m *Manager) GetStats() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()

    stats := make(map[string]interface{})
    
    for name, pool := range m.pools {
        pool.mu.RLock()
        healthyCount := 0
        for _, proxy := range pool.Proxies {
            if proxy.Healthy {
                healthyCount++
            }
        }
        
        stats[name] = map[string]interface{}{
            "total":   len(pool.Proxies),
            "healthy": healthyCount,
            "type":    pool.Type,
        }
        pool.mu.RUnlock()
    }

    return stats
}
