package engagement

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrDuplicateLike     = errors.New("like already exists")
	ErrLikeNotFound      = errors.New("like not found")
	ErrDuplicateFavorite = errors.New("favorite already exists")
	ErrFavoriteNotFound  = errors.New("favorite not found")
	ErrInvalidEngagement = errors.New("invalid engagement input")
	ErrUnauthorized      = errors.New("unauthorized")
)

type repository interface {
	CreateLike(ctx context.Context, item *Like) error
	DeleteLike(ctx context.Context, memberID, articleID uint) error
	CreateFavorite(ctx context.Context, item *Favorite) error
	DeleteFavorite(ctx context.Context, memberID, articleID uint) error
	ListFavorites(ctx context.Context, memberID uint, page, pageSize int) ([]Favorite, int64, error)
}

type Service struct {
	repo repository
}

type FavoritesResult struct {
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int64      `json:"total"`
	Items    []Favorite `json:"items"`
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Like(ctx context.Context, memberID, articleID uint) (*Like, error) {
	if s == nil || s.repo == nil || memberID == 0 || articleID == 0 {
		return nil, ErrInvalidEngagement
	}
	item := &Like{MemberID: memberID, ArticleID: articleID}
	if err := s.repo.CreateLike(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Unlike(ctx context.Context, memberID, articleID uint) error {
	if s == nil || s.repo == nil || memberID == 0 || articleID == 0 {
		return ErrInvalidEngagement
	}
	return s.repo.DeleteLike(ctx, memberID, articleID)
}

func (s *Service) Favorite(ctx context.Context, memberID, articleID uint) (*Favorite, error) {
	if s == nil || s.repo == nil || memberID == 0 || articleID == 0 {
		return nil, ErrInvalidEngagement
	}
	item := &Favorite{MemberID: memberID, ArticleID: articleID}
	if err := s.repo.CreateFavorite(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Unfavorite(ctx context.Context, memberID, articleID uint) error {
	if s == nil || s.repo == nil || memberID == 0 || articleID == 0 {
		return ErrInvalidEngagement
	}
	return s.repo.DeleteFavorite(ctx, memberID, articleID)
}

func (s *Service) ListFavorites(ctx context.Context, memberID uint, page, pageSize int) (*FavoritesResult, error) {
	if s == nil || s.repo == nil || memberID == 0 {
		return nil, ErrInvalidEngagement
	}
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.ListFavorites(ctx, memberID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &FavoritesResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrDuplicateLike):
		return http.StatusConflict, "like already exists"
	case errors.Is(err, ErrDuplicateFavorite):
		return http.StatusConflict, "favorite already exists"
	case errors.Is(err, ErrLikeNotFound):
		return http.StatusNotFound, "like not found"
	case errors.Is(err, ErrFavoriteNotFound):
		return http.StatusNotFound, "favorite not found"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrInvalidEngagement):
		return http.StatusBadRequest, "invalid engagement input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
