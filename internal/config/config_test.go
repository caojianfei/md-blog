package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromYAMLFile(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeConfigFile(t, `
app:
  addr: ":9090"
  env: "production"
  data_dir: "./storage"
  session_secret: "yaml-secret"
  read_timeout: 30
  write_timeout: 35
database:
  driver: "mysql"
  mysql_dsn: "blog:secret@tcp(127.0.0.1:3306)/md_blog?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: false
  max_open_conns: 20
  max_idle_conns: 8
  conn_max_lifetime: 600
bootstrap_admin:
  username: "root"
  password: "strong-pass"
`)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config from yaml: %v", err)
	}

	if cfg.App.Addr != ":9090" {
		t.Fatalf("expected addr :9090, got %q", cfg.App.Addr)
	}
	if cfg.App.DataDir != "./storage" {
		t.Fatalf("expected data dir ./storage, got %q", cfg.App.DataDir)
	}
	if cfg.Database.Driver != "mysql" {
		t.Fatalf("expected mysql driver, got %q", cfg.Database.Driver)
	}
	if cfg.Database.MySQLDSN == "" {
		t.Fatal("expected mysql dsn to be loaded from yaml")
	}
	if cfg.Database.AutoMigrate {
		t.Fatal("expected auto migrate to be false")
	}
	if cfg.BootstrapAdmin.Username != "root" {
		t.Fatalf("expected admin username root, got %q", cfg.BootstrapAdmin.Username)
	}
}

func TestEnvironmentOverridesYAML(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeConfigFile(t, `
app:
  addr: ":9090"
  data_dir: "./yaml-data"
  session_secret: "yaml-secret"
database:
  driver: "sqlite"
  sqlite_path: "./yaml-data/blog.db"
bootstrap_admin:
  username: "yaml-admin"
`)

	t.Setenv("APP_ADDR", ":9191")
	t.Setenv("APP_DATA_DIR", "./env-data")
	t.Setenv("DB_SQLITE_PATH", "./env-data/override.db")
	t.Setenv("ADMIN_USERNAME", "env-admin")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config with env override: %v", err)
	}

	if cfg.App.Addr != ":9191" {
		t.Fatalf("expected env addr :9191, got %q", cfg.App.Addr)
	}
	if cfg.App.DataDir != "./env-data" {
		t.Fatalf("expected env data dir ./env-data, got %q", cfg.App.DataDir)
	}
	if cfg.Database.SQLitePath != "./env-data/override.db" {
		t.Fatalf("expected env sqlite path override, got %q", cfg.Database.SQLitePath)
	}
	if cfg.BootstrapAdmin.Username != "env-admin" {
		t.Fatalf("expected env admin username env-admin, got %q", cfg.BootstrapAdmin.Username)
	}
}

func TestLoadWithoutConfigFileUsesDefaults(t *testing.T) {
	resetConfigEnv(t)
	withWorkingDir(t, t.TempDir(), func() {
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("load config without file: %v", err)
		}

		if cfg.App.Addr != ":8080" {
			t.Fatalf("expected default addr :8080, got %q", cfg.App.Addr)
		}
		if cfg.App.DataDir != "./data" {
			t.Fatalf("expected default data dir ./data, got %q", cfg.App.DataDir)
		}
		if cfg.Database.SQLitePath != "./data/blog.db" {
			t.Fatalf("expected default sqlite path ./data/blog.db, got %q", cfg.Database.SQLitePath)
		}
	})
}

func TestLoadWithMissingConfigFileReturnsError(t *testing.T) {
	resetConfigEnv(t)

	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected missing config file to return error")
	}
	if !strings.Contains(err.Error(), "missing.yaml") {
		t.Fatalf("expected error to mention missing file, got %v", err)
	}
}

func TestValidateRequiresDriverSpecificFields(t *testing.T) {
	resetConfigEnv(t)

	configPath := writeConfigFile(t, `
app:
  session_secret: "yaml-secret"
database:
  driver: "mysql"
  mysql_dsn: ""
`)

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected empty mysql dsn to fail validation")
	}
	if !strings.Contains(err.Error(), "DB_MYSQL_DSN") {
		t.Fatalf("expected mysql dsn validation error, got %v", err)
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	return path
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory to %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatalf("restore working directory to %s: %v", currentDir, err)
		}
	}()

	fn()
}

func resetConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_ADDR",
		"APP_ENV",
		"APP_DATA_DIR",
		"APP_SESSION_SECRET",
		"APP_READ_TIMEOUT",
		"APP_WRITE_TIMEOUT",
		"DB_DRIVER",
		"DB_SQLITE_PATH",
		"DB_MYSQL_DSN",
		"DB_AUTO_MIGRATE",
		"DB_MAX_OPEN_CONNS",
		"DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME",
		"ADMIN_USERNAME",
		"ADMIN_PASSWORD",
	} {
		t.Setenv(key, "")
	}
}
