package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	App            AppConfig
	Database       DatabaseConfig
	BootstrapAdmin AdminConfig
}

type AppConfig struct {
	Addr          string
	Env           string
	DataDir       string
	SessionSecret string
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

type AdminConfig struct {
	Username string
	Password string
}

func Load() Config {
	dataDir := getEnv("APP_DATA_DIR", "./data")

	cfg := Config{
		App: AppConfig{
			Addr:          getEnv("APP_ADDR", ":8080"),
			Env:           getEnv("APP_ENV", "development"),
			DataDir:       dataDir,
			SessionSecret: getEnv("APP_SESSION_SECRET", "change-me-session-secret"),
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
		BootstrapAdmin: AdminConfig{
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
	if c.Database.Driver != "sqlite" && c.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported DB_DRIVER: %s", c.Database.Driver)
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
