package category

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

func (r *Repository) Create(ctx context.Context, item *Category) error {
	if r == nil || r.db == nil {
		return ErrCategoryNotFound
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) List(ctx context.Context, page, pageSize int) ([]Category, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&Category{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]Category, 0)
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Order("id ASC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Category, error) {
	if r == nil || r.db == nil {
		return nil, ErrCategoryNotFound
	}

	var item Category
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Category, error) {
	if r == nil || r.db == nil {
		return nil, ErrCategoryNotFound
	}

	var item Category
	err := r.db.WithContext(ctx).Where("slug = ?", strings.TrimSpace(slug)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(ctx context.Context, item *Category) error {
	if r == nil || r.db == nil {
		return ErrCategoryNotFound
	}
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return ErrCategoryNotFound
	}

	result := r.db.WithContext(ctx).Delete(&Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
