package bootstrap

import (
	"io"
	"log/slog"
	"testing"

	"github.com/dovetaill/PureMux/pkg/config"
	"github.com/dovetaill/PureMux/pkg/database"
	"gorm.io/gorm"
)

type testBusinessModelOne struct{}
type testBusinessModelTwo struct{}

type testSchemaMigrator struct {
	models []any
}

func (m *testSchemaMigrator) AutoMigrate(models ...any) error {
	m.models = append(m.models, models...)
	return nil
}

func TestAutoMigrateRegistersAllBusinessModels(t *testing.T) {
	origModels := businessModels
	t.Cleanup(func() {
		businessModels = origModels
	})

	businessModels = nil
	RegisterBusinessModels(testBusinessModelOne{}, &testBusinessModelTwo{})

	migrator := &testSchemaMigrator{}
	if err := AutoMigrateBusinessTables(migrator); err != nil {
		t.Fatalf("AutoMigrateBusinessTables() error = %v", err)
	}

	if len(migrator.models) != 2 {
		t.Fatalf("migrated model count = %d, want %d", len(migrator.models), 2)
	}
	if _, ok := migrator.models[0].(testBusinessModelOne); !ok {
		t.Fatalf("first migrated model = %T, want %T", migrator.models[0], testBusinessModelOne{})
	}
	if _, ok := migrator.models[1].(*testBusinessModelTwo); !ok {
		t.Fatalf("second migrated model = %T, want %T", migrator.models[1], &testBusinessModelTwo{})
	}
}

func TestBuildServerRuntimeDoesNotAutoMigrateStarterSchema(t *testing.T) {
	origLoadConfig := loadConfigFn
	origNewLogger := newLoggerFn
	origBootstrapDatabase := bootstrapDatabaseFn
	origAutoMigrate := autoMigrateBusinessTablesFn
	origNewSeedStore := newSeedAdminStoreFn
	origHashPassword := seedAdminPasswordHashFn
	t.Cleanup(func() {
		loadConfigFn = origLoadConfig
		newLoggerFn = origNewLogger
		bootstrapDatabaseFn = origBootstrapDatabase
		autoMigrateBusinessTablesFn = origAutoMigrate
		newSeedAdminStoreFn = origNewSeedStore
		seedAdminPasswordHashFn = origHashPassword
	})

	resources := &database.Resources{MySQL: &gorm.DB{}}
	autoMigrateCalls := 0
	newSeedStoreCalls := 0
	hashCalls := 0

	loadConfigFn = func(path string) (*config.Config, error) {
		return &config.Config{App: config.AppConfig{Name: "PureMux"}}, nil
	}
	newLoggerFn = func(cfg config.LogConfig) (*slog.Logger, func() error, error) {
		return slog.New(slog.NewTextHandler(io.Discard, nil)), func() error { return nil }, nil
	}
	bootstrapDatabaseFn = func(cfg *config.Config) (*database.Resources, error) {
		return resources, nil
	}
	autoMigrateBusinessTablesFn = func(migrator schemaMigrator) error {
		autoMigrateCalls++
		return nil
	}
	newSeedAdminStoreFn = func(resources *database.Resources) SeedAdminStore {
		newSeedStoreCalls++
		return nil
	}
	seedAdminPasswordHashFn = func(password string) (string, error) {
		hashCalls++
		return "hashed:" + password, nil
	}

	_, err := BuildServerRuntime("configs/config.yaml")
	if err != nil {
		t.Fatalf("BuildServerRuntime() error = %v", err)
	}
	if autoMigrateCalls != 0 {
		t.Fatalf("auto migrate call count = %d, want %d", autoMigrateCalls, 0)
	}
	if newSeedStoreCalls != 0 {
		t.Fatalf("new seed admin store call count = %d, want %d", newSeedStoreCalls, 0)
	}
	if hashCalls != 0 {
		t.Fatalf("seed admin password hasher call count = %d, want %d", hashCalls, 0)
	}
}
