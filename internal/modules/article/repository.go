package article

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

func (r *Repository) Create(ctx context.Context, item *Article) error {
	if r == nil || r.db == nil {
		return ErrArticleNotFound
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) List(ctx context.Context, filter ListFilter, page, pageSize int) ([]Article, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}

	query := r.db.WithContext(ctx).Model(&Article{})
	if filter.AuthorID != nil {
		query = query.Where("author_id = ?", *filter.AuthorID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]Article, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Article, error) {
	if r == nil || r.db == nil {
		return nil, ErrArticleNotFound
	}

	var item Article
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Article, error) {
	if r == nil || r.db == nil {
		return nil, ErrArticleNotFound
	}

	var item Article
	err := r.db.WithContext(ctx).Where("slug = ?", strings.TrimSpace(slug)).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrArticleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(ctx context.Context, item *Article) error {
	if r == nil || r.db == nil {
		return ErrArticleNotFound
	}
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	if r == nil || r.db == nil {
		return ErrArticleNotFound
	}

	result := r.db.WithContext(ctx).Delete(&Article{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrArticleNotFound
	}
	return nil
}
