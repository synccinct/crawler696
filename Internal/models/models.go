// internal/models/models.go
package models

import (
    "time"
)

type CrawlTask struct {
    ID          string            `json:"id" bson:"_id"`
    URL         string            `json:"url" bson:"url"`
    Method      string            `json:"method" bson:"method"`
    Headers     map[string]string `json:"headers" bson:"headers"`
    Priority    int               `json:"priority" bson:"priority"`
    MaxDepth    int               `json:"max_depth" bson:"max_depth"`
    CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
    ScheduledAt time.Time         `json:"scheduled_at" bson:"scheduled_at"`
    Status      string            `json:"status" bson:"status"`
    SessionID   string            `json:"session_id" bson:"session_id"`
}

type CrawlResult struct {
    TaskID    string        `json:"task_id" bson:"task_id"`
    URL       string        `json:"url" bson:"url"`
    WorkerID  string        `json:"worker_id" bson:"worker_id"`
    Success   bool          `json:"success" bson:"success"`
    Data      *CrawlData    `json:"data,omitempty" bson:"data,omitempty"`
    Error     string        `json:"error,omitempty" bson:"error,omitempty"`
    StartTime time.Time     `json:"start_time" bson:"start_time"`
    EndTime   time.Time     `json:"end_time" bson:"end_time"`
    Duration  time.Duration `json:"duration" bson:"duration"`
}

type CrawlData struct {
    URL         string            `json:"url" bson:"url"`
    StatusCode  int               `json:"status_code" bson:"status_code"`
    Headers     map[string]string `json:"headers" bson:"headers"`
    Content     string            `json:"content" bson:"content"`
    Links       []string          `json:"links" bson:"links"`
    Images      []string          `json:"images" bson:"images"`
    Metadata    map[string]interface{} `json:"metadata" bson:"metadata"`
    Timestamp   time.Time         `json:"timestamp" bson:"timestamp"`
}

type CrawlSession struct {
    ID          string            `json:"id" bson:"_id"`
    Name        string            `json:"name" bson:"name"`
    Description string            `json:"description" bson:"description"`
    StartURLs   []string          `json:"start_urls" bson:"start_urls"`
    Rules       CrawlRules        `json:"rules" bson:"rules"`
    Status      string            `json:"status" bson:"status"`
    CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
    StartedAt   *time.Time        `json:"started_at,omitempty" bson:"started_at,omitempty"`
    CompletedAt *time.Time        `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
    Stats       SessionStats      `json:"stats" bson:"stats"`
}

type CrawlRules struct {
    MaxDepth        int      `json:"max_depth" bson:"max_depth"`
    MaxPages        int      `json:"max_pages" bson:"max_pages"`
    AllowedDomains  []string `json:"allowed_domains" bson:"allowed_domains"`
    BlockedDomains  []string `json:"blocked_domains" bson:"blocked_domains"`
    URLPatterns     []string `json:"url_patterns" bson:"url_patterns"`
    RespectRobotsTxt bool    `json:"respect_robots_txt" bson:"respect_robots_txt"`
    Delay           int      `json:"delay" bson:"delay"`
}

type SessionStats struct {
    TotalTasks      int `json:"total_tasks" bson:"total_tasks"`
    CompletedTasks  int `json:"completed_tasks" bson:"completed_tasks"`
    FailedTasks     int `json:"failed_tasks" bson:"failed_tasks"`
    PendingTasks    int `json:"pending_tasks" bson:"pending_tasks"`
    PagesPerMinute  int `json:"pages_per_minute" bson:"pages_per_minute"`
}

type ProxyInfo struct {
    ID          string    `json:"id" bson:"_id"`
    Host        string    `json:"host" bson:"host"`
    Port        int       `json:"port" bson:"port"`
    Username    string    `json:"username" bson:"username"`
    Password    string    `json:"password" bson:"password"`
    Type        string    `json:"type" bson:"type"`
    Country     string    `json:"country" bson:"country"`
    Provider    string    `json:"provider" bson:"provider"`
    Healthy     bool      `json:"healthy" bson:"healthy"`
    LastChecked time.Time `json:"last_checked" bson:"last_checked"`
    FailCount   int       `json:"fail_count" bson:"fail_count"`
}

type DetectionEvent struct {
    ID          string    `json:"id" bson:"_id"`
    URL         string    `json:"url" bson:"url"`
    ProxyID     string    `json:"proxy_id" bson:"proxy_id"`
    EventType   string    `json:"event_type" bson:"event_type"`
    Description string    `json:"description" bson:"description"`
    Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
    WorkerID    string    `json:"worker_id" bson:"worker_id"`
}
