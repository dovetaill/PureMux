package bootstrap

import (
	"context"
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

type testSeedAdminStore struct {
	hasAdmin       bool
	hasAdminCalls  int
	createCalls    int
	createdAccount SeedAdminAccount
}

func (s *testSeedAdminStore) HasAdmin(ctx context.Context) (bool, error) {
	s.hasAdminCalls++
	return s.hasAdmin, nil
}

func (s *testSeedAdminStore) CreateAdmin(ctx context.Context, account SeedAdminAccount) error {
	s.createCalls++
	s.createdAccount = account
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

func TestSeedAdminCreatesDefaultAdminWhenMissing(t *testing.T) {
	store := &testSeedAdminStore{}
	hashCalls := 0

	err := SeedDefaultAdmin(
		context.Background(),
		config.SeedAdminConfig{
			Enabled:  true,
			Username: "admin",
			Password: "admin123456",
		},
		store,
		func(password string) (string, error) {
			hashCalls++
			return "hashed:" + password, nil
		},
	)
	if err != nil {
		t.Fatalf("SeedDefaultAdmin() error = %v", err)
	}

	if store.hasAdminCalls != 1 {
		t.Fatalf("HasAdmin() call count = %d, want %d", store.hasAdminCalls, 1)
	}
	if hashCalls != 1 {
		t.Fatalf("password hasher call count = %d, want %d", hashCalls, 1)
	}
	if store.createCalls != 1 {
		t.Fatalf("CreateAdmin() call count = %d, want %d", store.createCalls, 1)
	}
	if store.createdAccount.Username != "admin" {
		t.Fatalf("created username = %q, want %q", store.createdAccount.Username, "admin")
	}
	if store.createdAccount.PasswordHash != "hashed:admin123456" {
		t.Fatalf("created password hash = %q, want %q", store.createdAccount.PasswordHash, "hashed:admin123456")
	}
	if store.createdAccount.Role != SeedAdminRoleAdmin {
		t.Fatalf("created role = %q, want %q", store.createdAccount.Role, SeedAdminRoleAdmin)
	}
	if store.createdAccount.Status != SeedAdminStatusActive {
		t.Fatalf("created status = %q, want %q", store.createdAccount.Status, SeedAdminStatusActive)
	}
}

func TestSeedAdminSkipsWhenDisabled(t *testing.T) {
	store := &testSeedAdminStore{}
	hashCalls := 0

	err := SeedDefaultAdmin(
		context.Background(),
		config.SeedAdminConfig{
			Enabled:  false,
			Username: "admin",
			Password: "admin123456",
		},
		store,
		func(password string) (string, error) {
			hashCalls++
			return "hashed:" + password, nil
		},
	)
	if err != nil {
		t.Fatalf("SeedDefaultAdmin() error = %v", err)
	}

	if store.hasAdminCalls != 0 {
		t.Fatalf("HasAdmin() call count = %d, want %d", store.hasAdminCalls, 0)
	}
	if store.createCalls != 0 {
		t.Fatalf("CreateAdmin() call count = %d, want %d", store.createCalls, 0)
	}
	if hashCalls != 0 {
		t.Fatalf("password hasher call count = %d, want %d", hashCalls, 0)
	}
}

func TestBuildServerRuntimeRunsSchemaWithoutSeedAdmin(t *testing.T) {
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

	cfg := &config.Config{
		App: config.AppConfig{Name: "PureMux"},
		Auth: config.AuthConfig{
			SeedAdmin: config.SeedAdminConfig{
				Enabled:  true,
				Username: "admin",
				Password: "admin123456",
			},
		},
	}
	resources := &database.Resources{MySQL: &gorm.DB{}}
	store := &testSeedAdminStore{}
	autoMigrateCalls := 0
	newSeedStoreCalls := 0
	hashCalls := 0

	loadConfigFn = func(path string) (*config.Config, error) {
		return cfg, nil
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
		return store
	}
	seedAdminPasswordHashFn = func(password string) (string, error) {
		hashCalls++
		return "hashed:" + password, nil
	}

	_, err := BuildServerRuntime("configs/config.yaml")
	if err != nil {
		t.Fatalf("BuildServerRuntime() error = %v", err)
	}

	if autoMigrateCalls != 1 {
		t.Fatalf("auto migrate call count = %d, want %d", autoMigrateCalls, 1)
	}
	if newSeedStoreCalls != 0 {
		t.Fatalf("new seed admin store call count = %d, want %d", newSeedStoreCalls, 0)
	}
	if store.createCalls != 0 {
		t.Fatalf("seed admin create count = %d, want %d", store.createCalls, 0)
	}
	if hashCalls != 0 {
		t.Fatalf("seed admin password hasher count = %d, want %d", hashCalls, 0)
	}
}
