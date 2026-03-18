package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dovetaill/PureMux/pkg/config"
)

func TestLoadReadsYAML(t *testing.T) {
	path := writeConfigFile(t, `
app:
  name: PureMux
  env: local
  host: 0.0.0.0
  port: 8080
mysql:
  host: 127.0.0.1
  port: 3306
  user: root
  password: root
  dbname: puremux
  charset: utf8mb4
  parse_time: true
  loc: Local
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime_minutes: 60
redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 2
log:
  level: info
  format: json
  output: both
  dir: logs
  filename: app.log
  max_size_mb: 100
  max_backups: 14
  max_age_days: 30
  compress: false
  rotate_daily: true
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "PureMux" {
		t.Fatalf("App.Name = %q, want %q", cfg.App.Name, "PureMux")
	}
	if cfg.App.Port != 8080 {
		t.Fatalf("App.Port = %d, want %d", cfg.App.Port, 8080)
	}
	if cfg.MySQL.DBName != "puremux" {
		t.Fatalf("MySQL.DBName = %q, want %q", cfg.MySQL.DBName, "puremux")
	}
	if cfg.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("Redis.Addr = %q, want %q", cfg.Redis.Addr, "127.0.0.1:6379")
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "info")
	}
}

func TestLoadEnvOverridesYAML(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")

	path := writeConfigFile(t, `
app:
  name: PureMux
mysql:
  host: 127.0.0.1
  user: root
  password: root
  dbname: puremux
redis:
  addr: 127.0.0.1:6379
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != 9090 {
		t.Fatalf("App.Port = %d, want %d", cfg.App.Port, 9090)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
}

func TestLoadReturnsErrorForMissingRequiredFields(t *testing.T) {
	path := writeConfigFile(t, `
app:
  env: local
mysql:
  host: 127.0.0.1
redis:
  addr: 127.0.0.1:6379
`)

	if _, err := config.Load(path); err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
