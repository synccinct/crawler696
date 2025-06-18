// pkg/storage/storage.go
package storage

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "crawler666/internal/models"

    _ "github.com/lib/pq"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/go-redis/redis/v8"
)

type Interface interface {
    StoreCrawlResult(result *models.CrawlResult) error
    GetPendingTasks(limit int) ([]*models.CrawlTask, error)
    CreateCrawlSession(session *models.CrawlSession) error
    UpdateSessionStats(sessionID string, stats *models.SessionStats) error
    GetCrawlSessions() ([]*models.CrawlSession, error)
    GetCrawlResults(sessionID string, limit int) ([]*models.CrawlResult, error)
    Close() error
}

type MultiStorage struct {
    postgres *PostgreSQLStorage
    mongodb  *MongoDBStorage
    redis    *RedisStorage
}

type PostgreSQLStorage struct {
    db *sql.DB
}

type MongoDBStorage struct {
    client   *mongo.Client
    database *mongo.Database
}

type RedisStorage struct {
    client *redis.Client
}

type Config struct {
    PostgreSQL PostgreSQLConfig
    MongoDB    MongoDBConfig
    Redis      RedisConfig
}

type PostgreSQLConfig struct {
    Host     string
    Port     int
    Database string
    Username string
    Password string
}

type MongoDBConfig struct {
    URI      string
    Database string
}

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

func NewMultiStorage(config Config) (*MultiStorage, error) {
    // Initialize PostgreSQL
    postgres, err := NewPostgreSQLStorage(config.PostgreSQL)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize PostgreSQL: %v", err)
    }

    // Initialize MongoDB
    mongodb, err := NewMongoDBStorage(config.MongoDB)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize MongoDB: %v", err)
    }

    // Initialize Redis
    redisStorage, err := NewRedisStorage(config.Redis)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Redis: %v", err)
    }

    return &MultiStorage{
        postgres: postgres,
        mongodb:  mongodb,
        redis:    redisStorage,
    }, nil
}

func NewPostgreSQLStorage(config PostgreSQLConfig) (*PostgreSQLStorage, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        config.Host, config.Port, config.Username, config.Password, config.Database)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    if err = db.Ping(); err != nil {
        return nil, err
    }

    storage := &PostgreSQLStorage{db: db}
    
    // Create tables
    if err := storage.createTables(); err != nil {
        return nil, err
    }

    return storage, nil
}

func (s *PostgreSQLStorage) createTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS crawl_sessions (
            id VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            description TEXT,
            start_urls TEXT[],
            rules JSONB,
            status VARCHAR(50),
            created_at TIMESTAMP DEFAULT NOW(),
            started_at TIMESTAMP,
            completed_at TIMESTAMP,
            stats JSONB
        )`,
        `CREATE TABLE IF NOT EXISTS crawl_tasks (
            id VARCHAR(255) PRIMARY KEY,
            session_id VARCHAR(255) REFERENCES crawl_sessions(id),
            url TEXT NOT NULL,
            method VARCHAR(10) DEFAULT 'GET',
            headers JSONB,
            priority INTEGER DEFAULT 0,
            max_depth INTEGER DEFAULT 0,
            created_at TIMESTAMP DEFAULT NOW(),
            scheduled_at TIMESTAMP,
            status VARCHAR(50) DEFAULT 'pending'
        )`,
        `CREATE TABLE IF NOT EXISTS proxy_info (
            id VARCHAR(255) PRIMARY KEY,
            host VARCHAR(255) NOT NULL,
            port INTEGER NOT NULL,
            username VARCHAR(255),
            password VARCHAR(255),
            type VARCHAR(50),
            country VARCHAR(50),
            provider VARCHAR(255),
            healthy BOOLEAN DEFAULT true,
            last_checked TIMESTAMP,
            fail_count INTEGER DEFAULT 0
        )`,
        `CREATE TABLE IF NOT EXISTS detection_events (
            id VARCHAR(255) PRIMARY KEY,
            url TEXT NOT NULL,
            proxy_id VARCHAR(255),
            event_type VARCHAR(100),
            description TEXT,
            timestamp TIMESTAMP DEFAULT NOW(),
            worker_id VARCHAR(255)
        )`,
        `CREATE INDEX IF NOT EXISTS idx_crawl_tasks_status ON crawl_tasks(status)`,
        `CREATE INDEX IF NOT EXISTS idx_crawl_tasks_session ON crawl_tasks(session_id)`,
        `CREATE INDEX IF NOT EXISTS idx_detection_events_timestamp ON detection_events(timestamp)`,
    }

    for _, query := range queries {
        if _, err := s.db.Exec(query); err != nil {
            return fmt.Errorf("failed to execute query: %v", err)
        }
    }

    return nil
}

func NewMongoDBStorage(config MongoDBConfig) (*MongoDBStorage, error) {
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.URI))
    if err != nil {
        return nil, err
    }

    database := client.Database(config.Database)

    return &MongoDBStorage{
        client:   client,
        database: database,
    }, nil
}

func NewRedisStorage(config RedisConfig) (*RedisStorage, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password: config.Password,
        DB:       config.DB,
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, err
    }

    return &RedisStorage{client: client}, nil
}

func (m *MultiStorage) StoreCrawlResult(result *models.CrawlResult) error {
    // Store in MongoDB for content
    if err := m.mongodb.StoreCrawlResult(result); err != nil {
        return err
    }

    // Cache in Redis for quick access
    return m.redis.CacheCrawlResult(result)
}

func (m *MultiStorage) GetPendingTasks(limit int) ([]*models.CrawlTask, error) {
    return m.postgres.GetPendingTasks(limit)
}

func (m *MultiStorage) CreateCrawlSession(session *models.CrawlSession) error {
    return m.postgres.CreateCrawlSession(session)
}

func (m *MultiStorage) UpdateSessionStats(sessionID string, stats *models.SessionStats) error {
    return m.postgres.UpdateSessionStats(sessionID, stats)
}

func (m *MultiStorage) GetCrawlSessions() ([]*models.CrawlSession, error) {
    return m.postgres.GetCrawlSessions()
}

func (m *MultiStorage) GetCrawlResults(sessionID string, limit int) ([]*models.CrawlResult, error) {
    return m.mongodb.GetCrawlResults(sessionID, limit)
}

func (s *PostgreSQLStorage) GetPendingTasks(limit int) ([]*models.CrawlTask, error) {
    query := `SELECT id, session_id, url, method, headers, priority, max_depth, 
              created_at, scheduled_at, status 
              FROM crawl_tasks 
              WHERE status = 'pending' 
              ORDER BY priority DESC, created_at ASC 
              LIMIT $1`

    rows, err := s.db.Query(query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []*models.CrawlTask
    for rows.Next() {
        task := &models.CrawlTask{}
        var headersJSON []byte

        err := rows.Scan(&task.ID, &task.SessionID, &task.URL, &task.Method,
            &headersJSON, &task.Priority, &task.MaxDepth, &task.CreatedAt,
            &task.ScheduledAt, &task.Status)
        if err != nil {
            return nil, err
        }

        if len(headersJSON) > 0 {
            json.Unmarshal(headersJSON, &task.Headers)
        }

        tasks = append(tasks, task)
    }

    return tasks, nil
}

func (s *PostgreSQLStorage) CreateCrawlSession(session *models.CrawlSession) error {
    rulesJSON, _ := json.Marshal(session.Rules)
    statsJSON, _ := json.Marshal(session.Stats)

    query := `INSERT INTO crawl_sessions (id, name, description, start_urls, rules, status, created_at, stats)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

    _, err := s.db.Exec(query, session.ID, session.Name, session.Description,
        fmt.Sprintf("{%s}", join(session.StartURLs, ",")),
        rulesJSON, session.Status, session.CreatedAt, statsJSON)

    return err
}

func (s *PostgreSQLStorage) UpdateSessionStats(sessionID string, stats *models.SessionStats) error {
    statsJSON, _ := json.Marshal(stats)
    query := `UPDATE crawl_sessions SET stats = $1 WHERE id = $2`
    _, err := s.db.Exec(query, statsJSON, sessionID)
    return err
}

func (s *PostgreSQLStorage) GetCrawlSessions() ([]*models.CrawlSession, error) {
    query := `SELECT id, name, description, start_urls, rules, status, 
              created_at, started_at, completed_at, stats 
              FROM crawl_sessions ORDER BY created_at DESC`

    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sessions []*models.CrawlSession
    for rows.Next() {
        session := &models.CrawlSession{}
        var rulesJSON, statsJSON []byte
        var startURLs string

        err := rows.Scan(&session.ID, &session.Name, &session.Description,
            &startURLs, &rulesJSON, &session.Status, &session.CreatedAt,
            &session.StartedAt, &session.CompletedAt, &statsJSON)
        if err != nil {
            return nil, err
        }

        // Parse start URLs (simplified)
        session.StartURLs = []string{startURLs}

        if len(rulesJSON) > 0 {
            json.Unmarshal(rulesJSON, &session.Rules)
        }
        if len(statsJSON) > 0 {
            json.Unmarshal(statsJSON, &session.Stats)
        }

        sessions = append(sessions, session)
    }

    return sessions, nil
}

func (m *MongoDBStorage) StoreCrawlResult(result *models.CrawlResult) error {
    collection := m.database.Collection("crawl_results")
    _, err := collection.InsertOne(context.Background(), result)
    return err
}

func (m *MongoDBStorage) GetCrawlResults(sessionID string, limit int) ([]*models.CrawlResult, error) {
    collection := m.database.Collection("crawl_results")
    
    filter := map[string]interface{}{}
    if sessionID != "" {
        filter["session_id"] = sessionID
    }

    opts := options.Find().SetLimit(int64(limit)).SetSort(map[string]int{"start_time": -1})
    cursor, err := collection.Find(context.Background(), filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.Background())

    var results []*models.CrawlResult
    for cursor.Next(context.Background()) {
        var result models.CrawlResult
        if err := cursor.Decode(&result); err != nil {
            return nil, err
        }
        results = append(results, &result)
    }

    return results, nil
}

func (r *RedisStorage) CacheCrawlResult(result *models.CrawlResult) error {
    key := fmt.Sprintf("result:%s", result.TaskID)
    data, _ := json.Marshal(result)
    return r.client.Set(context.Background(), key, data, time.Hour).Err()
}

func (m *MultiStorage) Close() error {
    if m.postgres != nil {
        m.postgres.db.Close()
    }
    if m.mongodb != nil {
        m.mongodb.client.Disconnect(context.Background())
    }
    if m.redis != nil {
        m.redis.client.Close()
    }
    return nil
}

func join(strs []string, sep string) string {
    if len(strs) == 0 {
        return ""
    }
    result := strs[0]
    for i := 1; i < len(strs); i++ {
        result += sep + strs[i]
    }
    return result
}
