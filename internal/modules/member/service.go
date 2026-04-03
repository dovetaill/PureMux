package member

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dovetaill/PureMux/internal/identity"
)

var (
	ErrMemberNotFound     = errors.New("member not found")
	ErrDuplicateMember    = errors.New("member already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidMemberInput = errors.New("invalid member input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrMemberDisabled     = errors.New("member disabled")
)

type repository interface {
	Create(ctx context.Context, item *Member) error
	FindByUsername(ctx context.Context, username string) (*Member, error)
	FindByID(ctx context.Context, id uint) (*Member, error)
}

type Service struct {
	repo   repository
	tokens *identity.TokenManager
}

type AuthResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Member    Profile   `json:"member"`
}

func NewService(repo repository, tokens *identity.TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, username, password string) (*AuthResult, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return nil, ErrUnauthorized
	}

	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || len(password) < 6 {
		return nil, ErrInvalidMemberInput
	}

	if existing, err := s.repo.FindByUsername(ctx, username); err == nil && existing != nil {
		return nil, ErrDuplicateMember
	} else if err != nil && !errors.Is(err, ErrMemberNotFound) {
		return nil, err
	}

	passwordHash, err := identity.HashPassword(password)
	if err != nil {
		return nil, err
	}

	item := &Member{
		Username:     username,
		PasswordHash: passwordHash,
		Status:       StatusActive,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	return s.issueToken(item)
}

func (s *Service) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return nil, ErrUnauthorized
	}

	item, err := s.repo.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if item.Status == StatusDisabled {
		return nil, ErrMemberDisabled
	}
	if err := identity.VerifyPassword(item.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueToken(item)
}

func (s *Service) Authenticate(ctx context.Context, token string) (*identity.Actor, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return nil, identity.ErrUnauthorized
	}

	claims, err := s.tokens.Parse(token)
	if err != nil {
		return nil, identity.ErrUnauthorized
	}

	actor, err := claims.Actor()
	if err != nil {
		return nil, err
	}
	if !actor.HasRole(RoleMember) {
		return nil, identity.ErrUnauthorized
	}

	item, err := s.repo.FindByID(ctx, actor.ID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return nil, identity.ErrUnauthorized
		}
		return nil, err
	}
	if item.Status == StatusDisabled {
		return nil, identity.ErrActorDisabled
	}
	if claims.Username != "" && claims.Username != item.Username {
		return nil, identity.ErrUnauthorized
	}

	verifiedActor := item.ToActor()
	return &verifiedActor, nil
}

func (s *Service) GetSelf(ctx context.Context, id uint) (*Profile, error) {
	if s == nil || s.repo == nil || id == 0 {
		return nil, ErrInvalidMemberInput
	}

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.Status == StatusDisabled {
		return nil, ErrMemberDisabled
	}

	profile := item.ToProfile()
	return &profile, nil
}

func (s *Service) issueToken(item *Member) (*AuthResult, error) {
	actor := item.ToActor()
	token, expiresAt, err := s.tokens.Sign(actor)
	if err != nil {
		return nil, err
	}
	return &AuthResult{
		Token:     token,
		ExpiresAt: expiresAt,
		Member:    item.ToProfile(),
	}, nil
}

func StatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrDuplicateMember):
		return http.StatusConflict, "member already exists"
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, ErrMemberDisabled):
		return http.StatusUnauthorized, "member disabled"
	case errors.Is(err, ErrMemberNotFound):
		return http.StatusNotFound, "member not found"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrInvalidMemberInput):
		return http.StatusBadRequest, "invalid member input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
