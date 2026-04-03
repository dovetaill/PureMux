package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dovetaill/PureMux/pkg/config"
)

const (
	SeedAdminRoleAdmin    = "admin"
	SeedAdminStatusActive = "active"
)

type schemaMigrator interface {
	AutoMigrate(dst ...any) error
}

type SeedAdminAccount struct {
	Username     string
	PasswordHash string
	Role         string
	Status       string
}

type SeedAdminStore interface {
	HasAdmin(ctx context.Context) (bool, error)
	CreateAdmin(ctx context.Context, account SeedAdminAccount) error
}

type passwordHasher func(password string) (string, error)

func RegisterBusinessModels(models ...any) {
	for _, model := range models {
		if model == nil {
			continue
		}
		businessModels = append(businessModels, model)
	}
}

func AutoMigrateBusinessTables(migrator schemaMigrator) error {
	if migrator == nil {
		return errors.New("schema migrator is required")
	}
	if len(businessModels) == 0 {
		return nil
	}
	if err := migrator.AutoMigrate(businessModels...); err != nil {
		return fmt.Errorf("auto migrate business models: %w", err)
	}
	return nil
}

func SeedDefaultAdmin(
	ctx context.Context,
	cfg config.SeedAdminConfig,
	store SeedAdminStore,
	hash passwordHasher,
) error {
	if !cfg.Enabled {
		return nil
	}
	if store == nil {
		return errors.New("seed admin store is required")
	}
	if hash == nil {
		return errors.New("seed admin password hasher is required")
	}

	username := strings.TrimSpace(cfg.Username)
	password := strings.TrimSpace(cfg.Password)
	if username == "" || password == "" {
		return errors.New("seed admin credentials are required")
	}

	hasAdmin, err := store.HasAdmin(ctx)
	if err != nil {
		return fmt.Errorf("check default admin: %w", err)
	}
	if hasAdmin {
		return nil
	}

	passwordHash, err := hash(password)
	if err != nil {
		return fmt.Errorf("hash default admin password: %w", err)
	}

	return store.CreateAdmin(ctx, SeedAdminAccount{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         SeedAdminRoleAdmin,
		Status:       SeedAdminStatusActive,
	})
}
