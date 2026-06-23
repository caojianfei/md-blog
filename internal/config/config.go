package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const defaultConfigPath = "./config.yaml"

type Config struct {
	App            AppConfig
	Database       DatabaseConfig
	BootstrapAdmin AdminConfig
	Turnstile      TurnstileConfig
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

type TurnstileConfig struct {
	SiteKey   string
	SecretKey string
}

func Load(configPath string) (Config, error) {
	resolvedPath, err := resolveConfigPath(configPath)
	if err != nil {
		return Config{}, err
	}

	reader, err := newConfigReader()
	if err != nil {
		return Config{}, err
	}
	if resolvedPath != "" {
		reader.SetConfigFile(resolvedPath)
		if err := reader.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config file %q: %w", resolvedPath, err)
		}
	}

	cfg := buildConfig(reader)
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.App.SessionSecret) == "" {
		return fmt.Errorf("APP_SESSION_SECRET/app.session_secret cannot be empty")
	}

	driver := strings.ToLower(strings.TrimSpace(c.Database.Driver))
	if driver != "sqlite" && driver != "mysql" {
		return fmt.Errorf("unsupported DB_DRIVER/database.driver: %s", c.Database.Driver)
	}
	if driver == "sqlite" && strings.TrimSpace(c.Database.SQLitePath) == "" {
		return fmt.Errorf("DB_SQLITE_PATH/database.sqlite_path cannot be empty when using sqlite")
	}
	if driver == "mysql" && strings.TrimSpace(c.Database.MySQLDSN) == "" {
		return fmt.Errorf("DB_MYSQL_DSN/database.mysql_dsn cannot be empty when using mysql")
	}
	return nil
}

func resolveConfigPath(configPath string) (string, error) {
	trimmedPath := strings.TrimSpace(configPath)
	if trimmedPath != "" {
		if _, err := os.Stat(trimmedPath); err != nil {
			return "", fmt.Errorf("config file %q: %w", trimmedPath, err)
		}
		return trimmedPath, nil
	}

	if _, err := os.Stat(defaultConfigPath); err == nil {
		return defaultConfigPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("check default config file %q: %w", defaultConfigPath, err)
	}
	return "", nil
}

func newConfigReader() (*viper.Viper, error) {
	reader := viper.New()
	reader.SetConfigType("yaml")
	reader.SetDefault("app.addr", ":8080")
	reader.SetDefault("app.env", "development")
	reader.SetDefault("app.data_dir", "./data")
	reader.SetDefault("app.session_secret", "change-me-session-secret")
	reader.SetDefault("app.read_timeout", 15)
	reader.SetDefault("app.write_timeout", 15)
	reader.SetDefault("database.driver", "sqlite")
	reader.SetDefault("database.mysql_dsn", "root:root@tcp(127.0.0.1:3306)/md_blog?charset=utf8mb4&parseTime=True&loc=Local")
	reader.SetDefault("database.auto_migrate", true)
	reader.SetDefault("database.max_open_conns", 10)
	reader.SetDefault("database.max_idle_conns", 5)
	reader.SetDefault("database.conn_max_lifetime", 300)
	reader.SetDefault("bootstrap_admin.username", "admin")
	reader.SetDefault("bootstrap_admin.password", "admin123456")
	reader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	reader.AutomaticEnv()

	envBindings := map[string]string{
		"app.addr":                   "APP_ADDR",
		"app.env":                    "APP_ENV",
		"app.data_dir":               "APP_DATA_DIR",
		"app.session_secret":         "APP_SESSION_SECRET",
		"app.read_timeout":           "APP_READ_TIMEOUT",
		"app.write_timeout":          "APP_WRITE_TIMEOUT",
		"database.driver":            "DB_DRIVER",
		"database.sqlite_path":       "DB_SQLITE_PATH",
		"database.mysql_dsn":         "DB_MYSQL_DSN",
		"database.auto_migrate":      "DB_AUTO_MIGRATE",
		"database.max_open_conns":    "DB_MAX_OPEN_CONNS",
		"database.max_idle_conns":    "DB_MAX_IDLE_CONNS",
		"database.conn_max_lifetime": "DB_CONN_MAX_LIFETIME",
		"bootstrap_admin.username":   "ADMIN_USERNAME",
		"bootstrap_admin.password":   "ADMIN_PASSWORD",
		"turnstile.site_key":         "TURNSTILE_SITE_KEY",
		"turnstile.secret_key":       "TURNSTILE_SECRET_KEY",
	}
	for key, envName := range envBindings {
		if err := reader.BindEnv(key, envName); err != nil {
			return nil, fmt.Errorf("bind env %s: %w", envName, err)
		}
	}
	return reader, nil
}

func buildConfig(reader *viper.Viper) Config {
	dataDir := strings.TrimSpace(reader.GetString("app.data_dir"))
	if dataDir == "" {
		dataDir = "./data"
	}

	sqlitePath := strings.TrimSpace(reader.GetString("database.sqlite_path"))
	if sqlitePath == "" {
		sqlitePath = defaultSQLitePath(dataDir)
	}

	return Config{
		App: AppConfig{
			Addr:          strings.TrimSpace(reader.GetString("app.addr")),
			Env:           strings.TrimSpace(reader.GetString("app.env")),
			DataDir:       dataDir,
			SessionSecret: strings.TrimSpace(reader.GetString("app.session_secret")),
			ReadTimeout:   reader.GetInt("app.read_timeout"),
			WriteTimeout:  reader.GetInt("app.write_timeout"),
		},
		Database: DatabaseConfig{
			Driver:          strings.ToLower(strings.TrimSpace(reader.GetString("database.driver"))),
			SQLitePath:      sqlitePath,
			MySQLDSN:        strings.TrimSpace(reader.GetString("database.mysql_dsn")),
			AutoMigrate:     reader.GetBool("database.auto_migrate"),
			MaxOpenConns:    reader.GetInt("database.max_open_conns"),
			MaxIdleConns:    reader.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: reader.GetInt("database.conn_max_lifetime"),
		},
		BootstrapAdmin: AdminConfig{
			Username: strings.TrimSpace(reader.GetString("bootstrap_admin.username")),
			Password: strings.TrimSpace(reader.GetString("bootstrap_admin.password")),
		},
		Turnstile: TurnstileConfig{
			SiteKey:   strings.TrimSpace(reader.GetString("turnstile.site_key")),
			SecretKey: strings.TrimSpace(reader.GetString("turnstile.secret_key")),
		},
	}
}

func defaultSQLitePath(dataDir string) string {
	cleanDataDir := strings.TrimSpace(dataDir)
	if cleanDataDir == "" {
		cleanDataDir = "./data"
	}
	return strings.TrimRight(cleanDataDir, "/") + "/blog.db"
}
