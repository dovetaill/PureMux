package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	if r == nil || r.db == nil {
		return nil, ErrUserNotFound
	}

	var user User
	err := r.db.WithContext(ctx).Where("username = ?", strings.TrimSpace(username)).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*User, error) {
	if r == nil || r.db == nil {
		return nil, ErrUserNotFound
	}

	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) HasAdmin(ctx context.Context) (bool, error) {
	if r == nil || r.db == nil {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("role = ?", RoleAdmin).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateAdmin(ctx context.Context, account bootstrap.SeedAdminAccount) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}

	return r.db.WithContext(ctx).Create(&User{
		Username:     account.Username,
		PasswordHash: account.PasswordHash,
		Role:         account.Role,
		Status:       account.Status,
	}).Error
}
