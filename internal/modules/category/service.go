package category

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrDuplicateSlug        = errors.New("slug already exists")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrInvalidCategoryInput = errors.New("invalid category input")
)

type repository interface {
	Create(ctx context.Context, item *Category) error
	List(ctx context.Context, page, pageSize int) ([]Category, int64, error)
	FindByID(ctx context.Context, id uint) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	Update(ctx context.Context, item *Category) error
	Delete(ctx context.Context, id uint) error
}

type Service struct {
	repo repository
}

type CreateInput struct {
	Name        string
	Slug        string
	Description string
}

type UpdateInput struct {
	ID          uint
	Name        string
	Slug        string
	Description *string
}

type ListResult struct {
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int64      `json:"total"`
	Items    []Category `json:"items"`
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Category, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidCategoryInput
	}

	name := strings.TrimSpace(input.Name)
	slug := normalizeSlug(input.Slug)
	if name == "" || slug == "" {
		return nil, ErrInvalidCategoryInput
	}

	if existing, err := s.repo.FindBySlug(ctx, slug); err == nil && existing != nil {
		return nil, ErrDuplicateSlug
	} else if err != nil && !errors.Is(err, ErrCategoryNotFound) {
		return nil, err
	}

	item := &Category{Name: name, Slug: slug, Description: strings.TrimSpace(input.Description)}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int) (*ListResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidCategoryInput
	}
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *Service) ListPublic(ctx context.Context, page, pageSize int) (*ListResult, error) {
	return s.List(ctx, page, pageSize)
}

func (s *Service) Get(ctx context.Context, id uint) (*Category, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrInvalidCategoryInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*Category, error) {
	if s == nil || s.repo == nil || input.ID == 0 {
		return nil, ErrInvalidCategoryInput
	}

	item, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if name := strings.TrimSpace(input.Name); name != "" {
		item.Name = name
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.Slug != "" {
		slug := normalizeSlug(input.Slug)
		if slug == "" {
			return nil, ErrInvalidCategoryInput
		}
		if slug != item.Slug {
			if existing, err := s.repo.FindBySlug(ctx, slug); err == nil && existing != nil && existing.ID != item.ID {
				return nil, ErrDuplicateSlug
			} else if err != nil && !errors.Is(err, ErrCategoryNotFound) {
				return nil, err
			}
			item.Slug = slug
		}
	}

	if strings.TrimSpace(item.Name) == "" || normalizeSlug(item.Slug) == "" {
		return nil, ErrInvalidCategoryInput
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s == nil || s.repo == nil || id == 0 {
		return ErrInvalidCategoryInput
	}
	return s.repo.Delete(ctx, id)
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrDuplicateSlug):
		return http.StatusConflict, "slug already exists"
	case errors.Is(err, ErrCategoryNotFound):
		return http.StatusNotFound, "category not found"
	case errors.Is(err, ErrInvalidCategoryInput):
		return http.StatusBadRequest, "invalid category input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
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
