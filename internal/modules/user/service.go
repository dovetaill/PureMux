package user

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

var (
	ErrDuplicateUsername = errors.New("username already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidUserInput  = errors.New("invalid user input")
)

type repository interface {
	Create(ctx context.Context, item *authmodule.User) error
	List(ctx context.Context, page, pageSize int) ([]authmodule.User, int64, error)
	FindByID(ctx context.Context, id uint) (*authmodule.User, error)
	FindByUsername(ctx context.Context, username string) (*authmodule.User, error)
	Update(ctx context.Context, item *authmodule.User) error
	Delete(ctx context.Context, id uint) error
}

type Service struct {
	repo       repository
	hashSecret func(password string) (string, error)
}

type CreateInput struct {
	Username string
	Password string
	Role     string
	Status   string
}

type UpdateInput struct {
	ID       uint
	Username string
	Password string
	Role     string
	Status   string
}

type ListResult struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
	Items    []authmodule.User `json:"items"`
}

func NewService(repo repository, hashSecret func(password string) (string, error)) *Service {
	return &Service{repo: repo, hashSecret: hashSecret}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*authmodule.User, error) {
	if s == nil || s.repo == nil || s.hashSecret == nil {
		return nil, ErrInvalidUserInput
	}

	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	role := normalizeRole(input.Role)
	status := normalizeStatus(input.Status)
	if username == "" || password == "" || role == "" || status == "" {
		return nil, ErrInvalidUserInput
	}

	if existing, err := s.repo.FindByUsername(ctx, username); err == nil && existing != nil {
		return nil, ErrDuplicateUsername
	} else if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hash, err := s.hashSecret(password)
	if err != nil {
		return nil, err
	}

	item := &authmodule.User{Username: username, PasswordHash: hash, Role: role, Status: status}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int) (*ListResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidUserInput
	}
	page, pageSize = normalizePage(page, pageSize)
	items, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListResult{Page: page, PageSize: pageSize, Total: total, Items: items}, nil
}

func (s *Service) Get(ctx context.Context, id uint) (*authmodule.User, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrInvalidUserInput
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*authmodule.User, error) {
	if s == nil || s.repo == nil || s.hashSecret == nil {
		return nil, ErrInvalidUserInput
	}
	if input.ID == 0 {
		return nil, ErrInvalidUserInput
	}

	item, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if username := strings.TrimSpace(input.Username); username != "" && username != item.Username {
		if existing, err := s.repo.FindByUsername(ctx, username); err == nil && existing != nil && existing.ID != item.ID {
			return nil, ErrDuplicateUsername
		} else if err != nil && !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		item.Username = username
	}
	if password := strings.TrimSpace(input.Password); password != "" {
		hash, err := s.hashSecret(password)
		if err != nil {
			return nil, err
		}
		item.PasswordHash = hash
	}
	if input.Role != "" {
		role := normalizeRole(input.Role)
		if role == "" {
			return nil, ErrInvalidUserInput
		}
		item.Role = role
	}
	if input.Status != "" {
		status := normalizeStatus(input.Status)
		if status == "" {
			return nil, ErrInvalidUserInput
		}
		item.Status = status
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if s == nil || s.repo == nil || id == 0 {
		return ErrInvalidUserInput
	}
	return s.repo.Delete(ctx, id)
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrDuplicateUsername):
		return http.StatusConflict, "username already exists"
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound, "user not found"
	case errors.Is(err, ErrInvalidUserInput):
		return http.StatusBadRequest, "invalid user input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleAdmin:
		return RoleAdmin
	case "", RoleUser:
		return RoleUser
	default:
		return ""
	}
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusActive, "":
		return StatusActive
	case StatusDisabled:
		return StatusDisabled
	default:
		return ""
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
