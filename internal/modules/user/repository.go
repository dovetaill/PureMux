package user

import (
	"context"
	"errors"
	"strings"

	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item *authmodule.User) error {
	if r == nil || r.db == nil {
		return ErrUserNotFound
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) List(ctx context.Context, page, pageSize int) ([]authmodule.User, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&authmodule.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]authmodule.User, 0)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("id ASC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*authmodule.User, error) {
	if r == nil || r.db == nil {
		return nil, ErrUserNotFound
	}

	var item authmodule.User
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*authmodule.User, error) {
	if r == nil || r.db == nil {
		return nil, ErrUserNotFound
	}

	var item authmodule.User
	err := r.db.WithContext(ctx).Where("username = ?", strings.TrimSpace(username)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(ctx context.Context, item *authmodule.User) error {
	if r == nil || r.db == nil {
		return ErrUserNotFound
	}
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return ErrUserNotFound
	}

	result := r.db.WithContext(ctx).Delete(&authmodule.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
