package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrUserDisabled       = errors.New("user disabled")
)

type userRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

type Service struct {
	repo   userRepository
	tokens *TokenManager
}

type LoginResult struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      CurrentUser `json:"user"`
}

func NewService(repo userRepository, tokens *TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return nil, ErrUnauthorized
	}

	user, err := s.repo.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status == StatusDisabled {
		return nil, ErrUserDisabled
	}
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	currentUser := user.ToCurrentUser()
	token, expiresAt, err := s.tokens.Sign(currentUser)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, ExpiresAt: expiresAt, User: currentUser}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*CurrentUser, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return nil, ErrUnauthorized
	}

	claims, err := s.tokens.Parse(token)
	if err != nil {
		return nil, ErrUnauthorized
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return nil, ErrUnauthorized
	}

	user, err := s.repo.FindByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if user.Status == StatusDisabled {
		return nil, ErrUserDisabled
	}

	currentUser := user.ToCurrentUser()
	if claims.Role != "" && claims.Role != currentUser.Role {
		return nil, ErrUnauthorized
	}
	if claims.Username != "" && claims.Username != currentUser.Username {
		return nil, ErrUnauthorized
	}
	return &currentUser, nil
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, ErrUserDisabled):
		return http.StatusUnauthorized, "user disabled"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "forbidden"
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrUserNotFound):
		return http.StatusUnauthorized, "unauthorized"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
