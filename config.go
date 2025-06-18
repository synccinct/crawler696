// config.go
package main

import (
    "os"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Crawler  CrawlerConfig  `yaml:"crawler"`
    Storage  StorageConfig  `yaml:"storage"`
    Proxy    ProxyConfig    `yaml:"proxy"`
    Stealth  StealthConfig  `yaml:"stealth"`
}

type ServerConfig struct {
    Port string `yaml:"port"`
    Host string `yaml:"host"`
}

type CrawlerConfig struct {
    MaxWorkers  int `yaml:"max_workers"`
    QueueSize   int `yaml:"queue_size"`
    RateLimit   int `yaml:"rate_limit"`
    UserAgent   string `yaml:"user_agent"`
    Timeout     int `yaml:"timeout"`
}

type StorageConfig struct {
    PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
    MongoDB    MongoDBConfig    `yaml:"mongodb"`
    Redis      RedisConfig      `yaml:"redis"`
}

type PostgreSQLConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type MongoDBConfig struct {
    URI      string `yaml:"uri"`
    Database string `yaml:"database"`
}

type RedisConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type ProxyConfig struct {
    Enabled     bool              `yaml:"enabled"`
    Pools       []ProxyPoolConfig `yaml:"pools"`
    Rotation    int               `yaml:"rotation_interval"`
    HealthCheck int               `yaml:"health_check_interval"`
}

type ProxyPoolConfig struct {
    Name      string   `yaml:"name"`
    Type      string   `yaml:"type"`
    Providers []string `yaml:"providers"`
    Endpoints []string `yaml:"endpoints"`
}

type StealthConfig struct {
    Enabled              bool `yaml:"enabled"`
    FingerprintRotation  bool `yaml:"fingerprint_rotation"`
    CanvasNoise          bool `yaml:"canvas_noise"`
    WebGLSpoofing        bool `yaml:"webgl_spoofing"`
    UserAgentRotation    bool `yaml:"user_agent_rotation"`
}

func LoadConfig(path string) (*Config, error) {
    config := &Config{
        Server: ServerConfig{
            Port: "8080",
            Host: "0.0.0.0",
        },
        Crawler: CrawlerConfig{
            MaxWorkers: 1000,
            QueueSize:  10000,
            RateLimit:  1000,
            UserAgent:  "Crawler666/1.0",
            Timeout:    30,
        },
    }

    if _, err := os.Stat(path); os.IsNotExist(err) {
        return config, nil // Return default config
    }

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    if err := yaml.Unmarshal(data, config); err != nil {
        return nil, err
    }

    return config, nil
}
