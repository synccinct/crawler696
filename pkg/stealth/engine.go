// pkg/stealth/engine.go
package stealth

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"
    "crawler666/pkg/proxy"
)

type Engine struct {
    config      *Config
    userAgents  []string
    profiles    map[string]*Profile
}

type Config struct {
    Enabled              bool
    FingerprintRotation  bool
    CanvasNoise          bool
    WebGLSpoofing        bool
    UserAgentRotation    bool
}

type Profile struct {
    UserAgent    string
    Viewport     Viewport
    Canvas       CanvasFingerprint
    WebGL        WebGLFingerprint
    Fonts        []string
    Timezone     string
    Language     string
    Platform     string
}

type Viewport struct {
    Width  int
    Height int
}

type CanvasFingerprint struct {
    Noise     float64
    TextValue string
}

type WebGLFingerprint struct {
    Vendor   string
    Renderer string
}

func NewEngine(config *Config) (*Engine, error) {
    engine := &Engine{
        config:   config,
        profiles: make(map[string]*Profile),
        userAgents: []string{
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
            // Add more user agents...
        },
    }

    return engine, nil
}

func (e *Engine) GenerateProfile(url string) (*Profile, error) {
    if !e.config.Enabled {
        return &Profile{}, nil
    }

    // Generate or retrieve cached profile
    profile := &Profile{
        UserAgent: e.selectRandomUserAgent(),
        Viewport:  e.generateRandomViewport(),
        Canvas:    e.generateCanvasFingerprint(),
        WebGL:     e.generateWebGLFingerprint(),
        Fonts:     e.generateFontList(),
        Timezone:  e.selectRandomTimezone(),
        Language:  "en-US,en;q=0.9",
        Platform:  e.selectRandomPlatform(),
    }

    return profile, nil
}

func (e *Engine) CreateHTTPClient(proxy *proxy.Proxy, profile *Profile) *http.Client {
    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    // Configure proxy if provided
    if proxy != nil {
        // Set up proxy transport
        // Implementation would configure HTTP proxy transport
    }

    return client
}

func (e *Engine) selectRandomUserAgent() string {
    if !e.config.UserAgentRotation || len(e.userAgents) == 0 {
        return "Crawler666/1.0"
    }
    return e.userAgents[rand.Intn(len(e.userAgents))]
}

func (e *Engine) generateRandomViewport() Viewport {
    viewports := []Viewport{
        {1920, 1080},
        {1366, 768},
        {1440, 900},
        {1536, 864},
        {1280, 720},
    }
    return viewports[rand.Intn(len(viewports))]
}

func (e *Engine) generateCanvasFingerprint() CanvasFingerprint {
    return CanvasFingerprint{
        Noise:     rand.Float64() * 0.1,
        TextValue: fmt.Sprintf("Crawler%d", rand.Intn(1000)),
    }
}

func (e *Engine) generateWebGLFingerprint() WebGLFingerprint {
    vendors := []string{"Google Inc.", "Mozilla", "Apple Inc."}
    renderers := []string{
        "ANGLE (Intel(R) HD Graphics 620 Direct3D11 vs_5_0 ps_5_0)",
        "WebKit WebGL",
        "Mozilla -- GPU",
    }
    
    return WebGLFingerprint{
        Vendor:   vendors[rand.Intn(len(vendors))],
        Renderer: renderers[rand.Intn(len(renderers))],
    }
}

func (e *Engine) generateFontList() []string {
    return []string{
        "Arial", "Helvetica", "Times New Roman", "Courier New",
        "Verdana", "Georgia", "Palatino", "Garamond",
    }
}

func (e *Engine) selectRandomTimezone() string {
    timezones := []string{
        "America/New_York", "America/Los_Angeles", "Europe/London",
        "Europe/Paris", "Asia/Tokyo", "Asia/Shanghai",
    }
    return timezones[rand.Intn(len(timezones))]
}

func (e *Engine) selectRandomPlatform() string {
    platforms := []string{"Win32", "MacIntel", "Linux x86_64"}
    return platforms[rand.Intn(len(platforms))]
}
