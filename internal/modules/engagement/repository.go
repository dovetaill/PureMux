package engagement

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateLike(ctx context.Context, item *Like) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateLike
		}
		return err
	}
	return nil
}

func (r *Repository) DeleteLike(ctx context.Context, memberID, articleID uint) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}
	result := r.db.WithContext(ctx).Where("member_id = ? AND article_id = ?", memberID, articleID).Delete(&Like{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLikeNotFound
	}
	return nil
}

func (r *Repository) CreateFavorite(ctx context.Context, item *Favorite) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateFavorite
		}
		return err
	}
	return nil
}

func (r *Repository) DeleteFavorite(ctx context.Context, memberID, articleID uint) error {
	if r == nil || r.db == nil {
		return ErrUnauthorized
	}
	result := r.db.WithContext(ctx).Where("member_id = ? AND article_id = ?", memberID, articleID).Delete(&Favorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFavoriteNotFound
	}
	return nil
}

func (r *Repository) ListFavorites(ctx context.Context, memberID uint, page, pageSize int) ([]Favorite, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, ErrUnauthorized
	}

	query := r.db.WithContext(ctx).Model(&Favorite{}).Where("member_id = ?", memberID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]Favorite, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("article_id ASC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func isDuplicateError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
