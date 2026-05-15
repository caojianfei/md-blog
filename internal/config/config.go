package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
)

type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Storage  StorageConfig
    Admin    AdminConfig
}

type AppConfig struct {
    Name          string
    Addr          string
    BaseURL       string
    Env           string
    DataDir       string
    SessionSecret string
    PreviewSecret string
    ReadTimeout   int
    WriteTimeout  int
}

type DatabaseConfig struct {
    Driver          string
    SQLitePath      string
    MySQLDSN        string
    AutoMigrate     bool
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime int
}

type StorageConfig struct {
    Driver        string
    LocalDir      string
    LocalBaseURL  string
    S3Endpoint    string
    S3Region      string
    S3Bucket      string
    S3AccessKey   string
    S3SecretKey   string
    S3UseSSL      bool
    S3PublicURL   string
    MaxUploadSize int64
}

type AdminConfig struct {
    Username string
    Password string
}

func Load() Config {
    dataDir := getEnv("APP_DATA_DIR", "./data")
    localUploadDir := getEnv("STORAGE_LOCAL_DIR", dataDir+"/uploads")

    cfg := Config{
        App: AppConfig{
            Name:          getEnv("APP_NAME", "Cybernote Blog"),
            Addr:          getEnv("APP_ADDR", ":8080"),
            BaseURL:       strings.TrimRight(getEnv("APP_BASE_URL", "http://localhost:8080"), "/"),
            Env:           getEnv("APP_ENV", "development"),
            DataDir:       dataDir,
            SessionSecret: getEnv("APP_SESSION_SECRET", "change-me-session-secret"),
            PreviewSecret: getEnv("APP_PREVIEW_SECRET", ""),
            ReadTimeout:   getIntEnv("APP_READ_TIMEOUT", 15),
            WriteTimeout:  getIntEnv("APP_WRITE_TIMEOUT", 15),
        },
        Database: DatabaseConfig{
            Driver:          strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
            SQLitePath:      getEnv("DB_SQLITE_PATH", dataDir+"/blog.db"),
            MySQLDSN:        getEnv("DB_MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/md_blog?charset=utf8mb4&parseTime=True&loc=Local"),
            AutoMigrate:     getBoolEnv("DB_AUTO_MIGRATE", true),
            MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 10),
            MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: getIntEnv("DB_CONN_MAX_LIFETIME", 300),
        },
        Storage: StorageConfig{
            Driver:        strings.ToLower(getEnv("STORAGE_DRIVER", "local")),
            LocalDir:      localUploadDir,
            LocalBaseURL:  strings.TrimRight(getEnv("STORAGE_LOCAL_BASE_URL", "/uploads"), "/"),
            S3Endpoint:    getEnv("STORAGE_S3_ENDPOINT", "127.0.0.1:9000"),
            S3Region:      getEnv("STORAGE_S3_REGION", "us-east-1"),
            S3Bucket:      getEnv("STORAGE_S3_BUCKET", "md-blog"),
            S3AccessKey:   getEnv("STORAGE_S3_ACCESS_KEY", ""),
            S3SecretKey:   getEnv("STORAGE_S3_SECRET_KEY", ""),
            S3UseSSL:      getBoolEnv("STORAGE_S3_USE_SSL", false),
            S3PublicURL:   strings.TrimRight(getEnv("STORAGE_S3_PUBLIC_URL", ""), "/"),
            MaxUploadSize: getInt64Env("STORAGE_MAX_UPLOAD_SIZE", 8<<20),
        },
        Admin: AdminConfig{
            Username: getEnv("ADMIN_USERNAME", "admin"),
            Password: getEnv("ADMIN_PASSWORD", "admin123456"),
        },
    }

    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    return cfg
}

func (c Config) Validate() error {
    if c.App.SessionSecret == "" {
        return fmt.Errorf("APP_SESSION_SECRET cannot be empty")
    }
    if c.App.PreviewSecret == "" {
        return fmt.Errorf("APP_PREVIEW_SECRET cannot be empty")
    }
    if c.Database.Driver != "sqlite" && c.Database.Driver != "mysql" {
        return fmt.Errorf("unsupported DB_DRIVER: %s", c.Database.Driver)
    }
    if c.Storage.Driver != "local" && c.Storage.Driver != "s3" {
        return fmt.Errorf("unsupported STORAGE_DRIVER: %s", c.Storage.Driver)
    }
    return nil
}

func getEnv(key, fallback string) string {
    value := strings.TrimSpace(os.Getenv(key))
    if value == "" {
        return fallback
    }
    return value
}

func getIntEnv(key string, fallback int) int {
    value := strings.TrimSpace(os.Getenv(key))
    if value == "" {
        return fallback
    }
    parsed, err := strconv.Atoi(value)
    if err != nil {
        return fallback
    }
    return parsed
}

func getInt64Env(key string, fallback int64) int64 {
    value := strings.TrimSpace(os.Getenv(key))
    if value == "" {
        return fallback
    }
    parsed, err := strconv.ParseInt(value, 10, 64)
    if err != nil {
        return fallback
    }
    return parsed
}

func getBoolEnv(key string, fallback bool) bool {
    value := strings.TrimSpace(os.Getenv(key))
    if value == "" {
        return fallback
    }
    parsed, err := strconv.ParseBool(value)
    if err != nil {
        return fallback
    }
    return parsed
}
