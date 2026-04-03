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

func TestLoadReadsDatabaseDriver(t *testing.T) {
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
database:
  driver: postgres
  mysql:
    host: 127.0.0.1
    user: root
    password: root
    dbname: puremux
  postgres:
    host: 127.0.0.1
    user: pg
    password: pg
    dbname: puremux
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Driver != "postgres" {
		t.Fatalf("Database.Driver = %q, want %q", cfg.Database.Driver, "postgres")
	}
}

func TestLoadReadsPostgresConfig(t *testing.T) {
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
database:
  driver: postgres
  mysql:
    host: 127.0.0.1
    user: root
    password: root
    dbname: puremux
  postgres:
    host: 10.0.0.1
    port: 5433
    user: pg_user
    password: pg_pwd
    dbname: puremux_db
    ssl_mode: disable
    time_zone: Asia/Shanghai
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Postgres.Host != "10.0.0.1" {
		t.Fatalf("Database.Postgres.Host = %q, want %q", cfg.Database.Postgres.Host, "10.0.0.1")
	}
	if cfg.Database.Postgres.Port != 5433 {
		t.Fatalf("Database.Postgres.Port = %d, want %d", cfg.Database.Postgres.Port, 5433)
	}
	if cfg.Database.Postgres.TimeZone != "Asia/Shanghai" {
		t.Fatalf("Database.Postgres.TimeZone = %q, want %q", cfg.Database.Postgres.TimeZone, "Asia/Shanghai")
	}
}

func TestLoadReadsQueueAndSchedulerConfig(t *testing.T) {
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
queue:
  asynq:
    concurrency: 15
    queue_name: default
scheduler:
  enabled: true
  spec: "@every 1m"
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Queue.Asynq.Concurrency != 15 {
		t.Fatalf("Queue.Asynq.Concurrency = %d, want %d", cfg.Queue.Asynq.Concurrency, 15)
	}
	if !cfg.Scheduler.Enabled {
		t.Fatal("Scheduler.Enabled = false, want true")
	}
	if cfg.Scheduler.Spec != "@every 1m" {
		t.Fatalf("Scheduler.Spec = %q, want %q", cfg.Scheduler.Spec, "@every 1m")
	}
}

func TestLoadReadsJWTConfig(t *testing.T) {
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
auth:
  jwt:
    secret: puremux-secret
    issuer: puremux-admin
    ttl_minutes: 180
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.JWT.Secret != "puremux-secret" {
		t.Fatalf("Auth.JWT.Secret = %q, want %q", cfg.Auth.JWT.Secret, "puremux-secret")
	}
	if cfg.Auth.JWT.Issuer != "puremux-admin" {
		t.Fatalf("Auth.JWT.Issuer = %q, want %q", cfg.Auth.JWT.Issuer, "puremux-admin")
	}
	if cfg.Auth.JWT.TTLMinutes != 180 {
		t.Fatalf("Auth.JWT.TTLMinutes = %d, want %d", cfg.Auth.JWT.TTLMinutes, 180)
	}
}

func TestLoadReadsSeedAdminConfig(t *testing.T) {
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
auth:
  seed_admin:
    enabled: true
    username: admin
    password: admin123456
log:
  level: info
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Auth.SeedAdmin.Enabled {
		t.Fatal("Auth.SeedAdmin.Enabled = false, want true")
	}
	if cfg.Auth.SeedAdmin.Username != "admin" {
		t.Fatalf("Auth.SeedAdmin.Username = %q, want %q", cfg.Auth.SeedAdmin.Username, "admin")
	}
	if cfg.Auth.SeedAdmin.Password != "admin123456" {
		t.Fatalf("Auth.SeedAdmin.Password = %q, want %q", cfg.Auth.SeedAdmin.Password, "admin123456")
	}
}

func TestLoadAppliesHTTPTimeoutDefaults(t *testing.T) {
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

	if cfg.HTTP.ReadTimeoutSeconds != 15 {
		t.Fatalf("HTTP.ReadTimeoutSeconds = %d, want %d", cfg.HTTP.ReadTimeoutSeconds, 15)
	}
	if cfg.HTTP.WriteTimeoutSeconds != 15 {
		t.Fatalf("HTTP.WriteTimeoutSeconds = %d, want %d", cfg.HTTP.WriteTimeoutSeconds, 15)
	}
	if cfg.HTTP.IdleTimeoutSeconds != 60 {
		t.Fatalf("HTTP.IdleTimeoutSeconds = %d, want %d", cfg.HTTP.IdleTimeoutSeconds, 60)
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
