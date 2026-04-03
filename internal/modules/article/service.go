package article

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

var (
	ErrArticleNotFound     = errors.New("article not found")
	ErrInvalidArticleInput = errors.New("invalid article input")
	ErrForbidden           = errors.New("forbidden")
)

type repository interface {
	Create(ctx context.Context, item *Article) error
	List(ctx context.Context, filter ListFilter, page, pageSize int) ([]Article, int64, error)
	FindByID(ctx context.Context, id uint) (*Article, error)
	Update(ctx context.Context, item *Article) error
	Delete(ctx context.Context, id uint) error
}

type Service struct {
	repo repository
	now  func() time.Time
}

type CreateInput struct {
	Title      string
	Summary    string
	Content    string
	CategoryID uint
}

type UpdateInput struct {
	ID         uint
	Title      string
	Summary    *string
	Content    *string
	CategoryID *uint
}

type ListResult struct {
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int64     `json:"total"`
	Items    []Article `json:"items"`
}

func NewService(repo repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) Create(ctx context.Context, actor authmodule.CurrentUser, input CreateInput) (*Article, error) {
	if s == nil || s.repo == nil || actor.ID == 0 {
		return nil, ErrInvalidArticleInput
	}

	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if title == "" || content == "" || input.CategoryID == 0 {
		return nil, ErrInvalidArticleInput
	}

	item := &Article{
		Title:      title,
		Summary:    strings.TrimSpace(input.Summary),
		Content:    content,
		Status:     StatusDraft,
		AuthorID:   actor.ID,
		CategoryID: input.CategoryID,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, actor authmodule.CurrentUser, page, pageSize int) (*ListResult, error) {
	if s == nil || s.repo == nil || actor.ID == 0 {
		return nil, ErrInvalidArticleInput
	}
	page, pageSize = normalizePage(page, pageSize)

	filter := ListFilter{}
	if actor.Role != authmodule.RoleAdmin {
		filter.AuthorID = &actor.ID
	}

	items, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *Service) ListPublic(ctx context.Context, page, pageSize int) (*ListResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidArticleInput
	}
	page, pageSize = normalizePage(page, pageSize)

	status := StatusPublished
	items, total, err := s.repo.List(ctx, ListFilter{Status: &status}, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *Service) Get(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error) {
	item, err := s.loadOwnedArticle(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, actor authmodule.CurrentUser, input UpdateInput) (*Article, error) {
	if input.ID == 0 {
		return nil, ErrInvalidArticleInput
	}
	item, err := s.loadOwnedArticle(ctx, actor, input.ID)
	if err != nil {
		return nil, err
	}

	if title := strings.TrimSpace(input.Title); title != "" {
		item.Title = title
	}
	if input.Summary != nil {
		item.Summary = strings.TrimSpace(*input.Summary)
	}
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return nil, ErrInvalidArticleInput
		}
		item.Content = content
	}
	if input.CategoryID != nil {
		if *input.CategoryID == 0 {
			return nil, ErrInvalidArticleInput
		}
		item.CategoryID = *input.CategoryID
	}
	if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Content) == "" || item.CategoryID == 0 {
		return nil, ErrInvalidArticleInput
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, actor authmodule.CurrentUser, id uint) error {
	if id == 0 {
		return ErrInvalidArticleInput
	}
	if _, err := s.loadOwnedArticle(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Publish(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error) {
	if id == 0 {
		return nil, ErrInvalidArticleInput
	}
	item, err := s.loadOwnedArticle(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	now := s.now()
	item.Status = StatusPublished
	item.PublishedAt = &now
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Unpublish(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error) {
	if id == 0 {
		return nil, ErrInvalidArticleInput
	}
	item, err := s.loadOwnedArticle(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	item.Status = StatusDraft
	item.PublishedAt = nil
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) loadOwnedArticle(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error) {
	if s == nil || s.repo == nil || actor.ID == 0 || id == 0 {
		return nil, ErrInvalidArticleInput
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Role == authmodule.RoleAdmin || item.AuthorID == actor.ID {
		return item, nil
	}
	return nil, ErrForbidden
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, ErrArticleNotFound):
		return http.StatusNotFound, "article not found"
	case errors.Is(err, ErrInvalidArticleInput):
		return http.StatusBadRequest, "invalid article input"
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
