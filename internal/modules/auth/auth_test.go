package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/middleware"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	"github.com/dovetaill/PureMux/pkg/config"
)

type authRepoStub struct {
	usersByUsername map[string]*authmodule.User
	usersByID       map[uint]*authmodule.User
}

func newAuthRepoStub(users ...*authmodule.User) *authRepoStub {
	stub := &authRepoStub{
		usersByUsername: make(map[string]*authmodule.User, len(users)),
		usersByID:       make(map[uint]*authmodule.User, len(users)),
	}
	for _, user := range users {
		stub.usersByUsername[user.Username] = user
		stub.usersByID[user.ID] = user
	}
	return stub
}

func (s *authRepoStub) FindByUsername(ctx context.Context, username string) (*authmodule.User, error) {
	if user, ok := s.usersByUsername[username]; ok {
		return user, nil
	}
	return nil, authmodule.ErrUserNotFound
}

func (s *authRepoStub) FindByID(ctx context.Context, id uint) (*authmodule.User, error) {
	if user, ok := s.usersByID[id]; ok {
		return user, nil
	}
	return nil, authmodule.ErrUserNotFound
}

func TestLoginReturnsJWTForValidCredentials(t *testing.T) {
	hash := mustHashPassword(t, "secret123")
	svc := newAuthService(t, &authmodule.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Role:         authmodule.RoleAdmin,
		Status:       authmodule.StatusActive,
	})

	handler := newAuthHandler(t, svc, false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if token, _ := data["token"].(string); token == "" {
		t.Fatal("token = empty, want non-empty")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	hash := mustHashPassword(t, "secret123")
	svc := newAuthService(t, &authmodule.User{
		ID:           1,
		Username:     "admin",
		PasswordHash: hash,
		Role:         authmodule.RoleAdmin,
		Status:       authmodule.StatusActive,
	})

	handler := newAuthHandler(t, svc, false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"bad-password"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "invalid credentials" {
		t.Fatalf("message = %q, want %q", got.Message, "invalid credentials")
	}
}

func TestAuthMiddlewareLoadsCurrentUserFromBearerToken(t *testing.T) {
	hash := mustHashPassword(t, "secret123")
	svc := newAuthService(t, &authmodule.User{
		ID:           9,
		Username:     "editor",
		PasswordHash: hash,
		Role:         authmodule.RoleUser,
		Status:       authmodule.StatusActive,
	})
	token := mustLogin(t, svc, "editor", "secret123")

	handler := newAuthHandler(t, svc, true)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["username"] != "editor" {
		t.Fatalf("username = %v, want %q", data["username"], "editor")
	}
	if data["role"] != authmodule.RoleUser {
		t.Fatalf("role = %v, want %q", data["role"], authmodule.RoleUser)
	}
}

func TestAuthMiddlewareRejectsDisabledUser(t *testing.T) {
	hash := mustHashPassword(t, "secret123")
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	user := &authmodule.User{
		ID:           12,
		Username:     "disabled-user",
		PasswordHash: hash,
		Role:         authmodule.RoleUser,
		Status:       authmodule.StatusDisabled,
	}
	token := mustSignToken(t, tokens, user)
	svc := authmodule.NewService(newAuthRepoStub(user), tokens)

	handler := newAuthHandler(t, svc, true)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "user disabled" {
		t.Fatalf("message = %q, want %q", got.Message, "user disabled")
	}
}

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	hash := mustHashPassword(t, "secret123")
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	user := &authmodule.User{
		ID:           15,
		Username:     "writer",
		PasswordHash: hash,
		Role:         authmodule.RoleUser,
		Status:       authmodule.StatusActive,
	}
	token := mustSignToken(t, tokens, user)
	svc := authmodule.NewService(newAuthRepoStub(user), tokens)

	handler := middleware.Authenticate(svc)(middleware.RequireAdmin()(middleware.RequireAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "forbidden" {
		t.Fatalf("message = %q, want %q", got.Message, "forbidden")
	}
}

func newAuthService(t *testing.T, users ...*authmodule.User) *authmodule.Service {
	t.Helper()
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	return authmodule.NewService(newAuthRepoStub(users...), tokens)
}

func newAuthHandler(t *testing.T, svc *authmodule.Service, requireAuth bool) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	authmodule.RegisterRoutes(api, svc)

	if requireAuth {
		return middleware.Authenticate(svc)(middleware.RequireAuthenticated()(mux))
	}
	return middleware.Authenticate(svc)(mux)
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := authmodule.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	return hash
}

func mustLogin(t *testing.T, svc *authmodule.Service, username, password string) string {
	t.Helper()
	result, err := svc.Login(context.Background(), username, password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	return result.Token
}

func mustSignToken(t *testing.T, tokens *authmodule.TokenManager, user *authmodule.User) string {
	t.Helper()
	token, _, err := tokens.Sign(user.ToCurrentUser())
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	return token
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

func envelopeData(t *testing.T, envelope response.Envelope) map[string]any {
	t.Helper()
	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", envelope.Data)
	}
	return data
}
