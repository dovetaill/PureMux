package member

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, item *Member) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*Member, error) {
	if r == nil || r.db == nil {
		return nil, ErrMemberNotFound
	}

	var item Member
	err := r.db.WithContext(ctx).Where("username = ?", strings.TrimSpace(username)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Member, error) {
	if r == nil || r.db == nil {
		return nil, ErrMemberNotFound
	}

	var item Member
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
