package database

import (
	"fmt"
	"strings"

	"github.com/dovetaill/PureMux/pkg/config"
	"gorm.io/gorm"
)

const (
	driverMySQL    = "mysql"
	driverPostgres = "postgres"
)

func openPrimaryDatabase(cfg *config.Config) (*gorm.DB, error) {
	driver := normalizedDriver(cfg.Database.Driver)
	switch driver {
	case driverMySQL:
		return openMySQLFn(resolveMySQLConfig(cfg))
	case driverPostgres:
		return openPostgresFn(cfg.Database.Postgres)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
}

func normalizedDriver(driver string) string {
	if strings.TrimSpace(driver) == "" {
		return driverMySQL
	}
	return strings.ToLower(strings.TrimSpace(driver))
}

func resolveMySQLConfig(cfg *config.Config) config.MySQLConfig {
	if cfg.Database.MySQL.Host != "" || cfg.Database.MySQL.User != "" || cfg.Database.MySQL.DBName != "" {
		return config.MySQLConfig{
			Host:                   cfg.Database.MySQL.Host,
			Port:                   cfg.Database.MySQL.Port,
			User:                   cfg.Database.MySQL.User,
			Password:               cfg.Database.MySQL.Password,
			DBName:                 cfg.Database.MySQL.DBName,
			Charset:                cfg.Database.MySQL.Charset,
			ParseTime:              cfg.Database.MySQL.ParseTime,
			Loc:                    cfg.Database.MySQL.Loc,
			MaxOpenConns:           cfg.Database.MySQL.MaxOpenConns,
			MaxIdleConns:           cfg.Database.MySQL.MaxIdleConns,
			ConnMaxLifetimeMinutes: cfg.Database.MySQL.ConnMaxLifetimeMinutes,
		}
	}

	return cfg.MySQL
}
