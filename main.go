// main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "crawler666/internal/models"
    "crawler666/pkg/proxy"
    "crawler666/pkg/stealth"
    "crawler666/pkg/storage"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

type CrawlerApp struct {
    Engine      *CrawlerEngine
    ProxyMgr    *proxy.Manager
    StealthEng  *stealth.Engine
    Storage     storage.Interface
    Config      *Config
    Logger      *logrus.Logger
}

func main() {
    // Initialize logger
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)

    // Load configuration
    config, err := LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize storage
    storage, err := storage.NewMultiStorage(config.Storage)
    if err != nil {
        log.Fatalf("Failed to initialize storage: %v", err)
    }

    // Initialize proxy manager
    proxyMgr, err := proxy.NewManager(config.Proxy)
    if err != nil {
        log.Fatalf("Failed to initialize proxy manager: %v", err)
    }

    // Initialize stealth engine
    stealthEng, err := stealth.NewEngine(config.Stealth)
    if err != nil {
        log.Fatalf("Failed to initialize stealth engine: %v", err)
    }

    // Initialize crawler engine
    crawlerEngine := NewCrawlerEngine(config.Crawler, storage, proxyMgr, stealthEng, logger)

    app := &CrawlerApp{
        Engine:     crawlerEngine,
        ProxyMgr:   proxyMgr,
        StealthEng: stealthEng,
        Storage:    storage,
        Config:     config,
        Logger:     logger,
    }

    // Start HTTP server
    router := setupRoutes(app)
    server := &http.Server{
        Addr:    ":" + config.Server.Port,
        Handler: router,
    }

    // Start crawler workers
    go app.Engine.StartWorkers(context.Background())

    // Start server
    go func() {
        logger.Infof("Starting Crawler666 server on port %s", config.Server.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Server failed to start: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down Crawler666...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Errorf("Server forced to shutdown: %v", err)
    }

    app.Engine.Stop()
    logger.Info("Crawler666 stopped")
}

func setupRoutes(app *CrawlerApp) *gin.Engine {
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    router.Use(gin.Logger(), gin.Recovery())

    // Static files for frontend
    router.Static("/static", "./web/dist")
    router.StaticFile("/", "./web/dist/index.html")

    // API routes
    api := router.Group("/api/v1")
    {
        // Crawler management
        api.POST("/crawl", app.startCrawl)
        api.GET("/crawl/:id", app.getCrawlStatus)
        api.DELETE("/crawl/:id", app.stopCrawl)
        api.GET("/crawls", app.listCrawls)

        // Configuration
        api.GET("/config", app.getConfig)
        api.PUT("/config", app.updateConfig)

        // Monitoring
        api.GET("/stats", app.getStats)
        api.GET("/health", app.healthCheck)
        api.GET("/metrics", app.getMetrics)

        // Proxy management
        api.GET("/proxies", app.getProxies)
        api.POST("/proxies/test", app.testProxy)

        // Data export
        api.GET("/export/:crawlId", app.exportData)
    }

    return router
}
