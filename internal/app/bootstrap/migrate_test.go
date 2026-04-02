package bootstrap

import (
	"strings"
	"testing"

	"github.com/dovetaill/PureMux/pkg/config"
)

func TestBuildMigrateConfigUsesSelectedDriver(t *testing.T) {
	t.Run("mysql", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{Driver: "mysql"},
			MySQL: config.MySQLConfig{
				Host:      "127.0.0.1",
				Port:      3306,
				User:      "root",
				Password:  "root",
				DBName:    "puremux",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Local",
			},
		}

		got, err := BuildMigrateConfig(cfg)
		if err != nil {
			t.Fatalf("BuildMigrateConfig() error = %v", err)
		}
		if got.Driver != "mysql" {
			t.Fatalf("Driver = %q, want %q", got.Driver, "mysql")
		}
		if got.SourceURL != "file://migrations" {
			t.Fatalf("SourceURL = %q, want %q", got.SourceURL, "file://migrations")
		}
		wantURL := "mysql://root:root@tcp(127.0.0.1:3306)/puremux?charset=utf8mb4&loc=Local&parseTime=true"
		if got.DatabaseURL != wantURL {
			t.Fatalf("DatabaseURL = %q, want %q", got.DatabaseURL, wantURL)
		}
	})

	t.Run("postgres", func(t *testing.T) {
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Driver: "postgres",
				Postgres: config.PostgresConfig{
					Host:     "127.0.0.1",
					Port:     5432,
					User:     "postgres",
					Password: "secret",
					DBName:   "puremux",
					SSLMode:  "disable",
					TimeZone: "Asia/Shanghai",
				},
			},
		}

		got, err := BuildMigrateConfig(cfg)
		if err != nil {
			t.Fatalf("BuildMigrateConfig() error = %v", err)
		}
		if got.Driver != "postgres" {
			t.Fatalf("Driver = %q, want %q", got.Driver, "postgres")
		}
		wantURL := "postgres://postgres:secret@127.0.0.1:5432/puremux?TimeZone=Asia%2FShanghai&sslmode=disable"
		if got.DatabaseURL != wantURL {
			t.Fatalf("DatabaseURL = %q, want %q", got.DatabaseURL, wantURL)
		}
	})
}

func TestMigrateCommandRejectsUnsupportedDriver(t *testing.T) {
	origLoadConfig := loadConfigFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
	})

	loadConfigFn = func(path string) (*config.Config, error) {
		return &config.Config{
			Database: config.DatabaseConfig{Driver: "sqlite"},
		}, nil
	}

	err := RunMigrateCommand("configs/config.yaml")
	if err == nil {
		t.Fatal("RunMigrateCommand() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Fatalf("error = %v, want contains %q", err, "unsupported database driver")
	}
}
