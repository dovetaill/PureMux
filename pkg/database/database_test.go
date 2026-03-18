package database

import (
	"testing"

	"github.com/dovetaill/PureMux/pkg/config"
)

func TestBuildMySQLDSN(t *testing.T) {
	cfg := config.MySQLConfig{
		Host:                   "127.0.0.1",
		Port:                   3306,
		User:                   "root",
		Password:               "root",
		DBName:                 "puremux",
		Charset:                "utf8mb4",
		ParseTime:              true,
		Loc:                    "Local",
		MaxOpenConns:           20,
		MaxIdleConns:           10,
		ConnMaxLifetimeMinutes: 60,
	}

	got := buildMySQLDSN(cfg)
	want := "root:root@tcp(127.0.0.1:3306)/puremux?charset=utf8mb4&parseTime=true&loc=Local"
	if got != want {
		t.Fatalf("buildMySQLDSN() = %q, want %q", got, want)
	}
}

func TestBuildRedisOptions(t *testing.T) {
	cfg := config.RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "secret",
		DB:           2,
		PoolSize:     32,
		MinIdleConns: 8,
	}

	opts := buildRedisOptions(cfg)
	if opts.Addr != cfg.Addr {
		t.Fatalf("Addr = %q, want %q", opts.Addr, cfg.Addr)
	}
	if opts.Password != cfg.Password {
		t.Fatalf("Password = %q, want %q", opts.Password, cfg.Password)
	}
	if opts.DB != cfg.DB {
		t.Fatalf("DB = %d, want %d", opts.DB, cfg.DB)
	}
	if opts.PoolSize != cfg.PoolSize {
		t.Fatalf("PoolSize = %d, want %d", opts.PoolSize, cfg.PoolSize)
	}
	if opts.MinIdleConns != cfg.MinIdleConns {
		t.Fatalf("MinIdleConns = %d, want %d", opts.MinIdleConns, cfg.MinIdleConns)
	}
}

func TestResourcesCloseIsSafe(t *testing.T) {
	var nilResources *Resources
	if err := nilResources.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}

	emptyResources := &Resources{}
	if err := emptyResources.Close(); err != nil {
		t.Fatalf("empty Close() error = %v", err)
	}
}
