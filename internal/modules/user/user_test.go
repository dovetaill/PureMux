package user_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/middleware"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	"github.com/dovetaill/PureMux/internal/modules/user"
	"github.com/dovetaill/PureMux/pkg/config"
)

type authRepoStub struct {
	usersByID map[uint]*authmodule.User
}

func newAuthRepoStub(users ...*authmodule.User) *authRepoStub {
	stub := &authRepoStub{usersByID: make(map[uint]*authmodule.User, len(users))}
	for _, item := range users {
		stub.usersByID[item.ID] = cloneAuthUser(item)
	}
	return stub
}

func (s *authRepoStub) FindByUsername(ctx context.Context, username string) (*authmodule.User, error) {
	for _, item := range s.usersByID {
		if item.Username == username {
			return cloneAuthUser(item), nil
		}
	}
	return nil, authmodule.ErrUserNotFound
}

func (s *authRepoStub) FindByID(ctx context.Context, id uint) (*authmodule.User, error) {
	item, ok := s.usersByID[id]
	if !ok {
		return nil, authmodule.ErrUserNotFound
	}
	return cloneAuthUser(item), nil
}

type userRepoStub struct {
	nextID uint
	items  map[uint]*authmodule.User
}

func newUserRepoStub(users ...*authmodule.User) *userRepoStub {
	stub := &userRepoStub{nextID: 1, items: make(map[uint]*authmodule.User, len(users))}
	for _, item := range users {
		clone := cloneAuthUser(item)
		if clone == nil {
			continue
		}
		stub.items[clone.ID] = clone
		if clone.ID >= stub.nextID {
			stub.nextID = clone.ID + 1
		}
	}
	return stub
}

func (s *userRepoStub) Create(ctx context.Context, item *authmodule.User) error {
	clone := cloneAuthUser(item)
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	s.items[clone.ID] = clone
	item.ID = clone.ID
	item.CreatedAt = clone.CreatedAt
	item.UpdatedAt = clone.UpdatedAt
	return nil
}

func (s *userRepoStub) List(ctx context.Context, page, pageSize int) ([]authmodule.User, int64, error) {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	items := make([]authmodule.User, 0, len(ids))
	for _, id := range ids {
		items = append(items, *cloneAuthUser(s.items[uint(id)]))
	}
	return items, int64(len(items)), nil
}

func (s *userRepoStub) FindByID(ctx context.Context, id uint) (*authmodule.User, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return cloneAuthUser(item), nil
}

func (s *userRepoStub) FindByUsername(ctx context.Context, username string) (*authmodule.User, error) {
	for _, item := range s.items {
		if item.Username == username {
			return cloneAuthUser(item), nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (s *userRepoStub) Update(ctx context.Context, item *authmodule.User) error {
	if _, ok := s.items[item.ID]; !ok {
		return user.ErrUserNotFound
	}
	clone := cloneAuthUser(item)
	clone.UpdatedAt = time.Now()
	s.items[item.ID] = clone
	item.UpdatedAt = clone.UpdatedAt
	return nil
}

func (s *userRepoStub) Delete(ctx context.Context, id uint) error {
	if _, ok := s.items[id]; !ok {
		return user.ErrUserNotFound
	}
	delete(s.items, id)
	return nil
}

func TestAdminCanCreateUser(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleAdmin, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(`{"username":"alice","password":"secret123","role":"user","status":"active"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["username"] != "alice" {
		t.Fatalf("username = %v, want %q", data["username"], "alice")
	}
	if data["role"] != authmodule.RoleUser {
		t.Fatalf("role = %v, want %q", data["role"], authmodule.RoleUser)
	}
}

func TestAdminCanListUsers(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleAdmin,
		&authmodule.User{ID: 2, Username: "alice", Role: authmodule.RoleUser, Status: authmodule.StatusActive},
		&authmodule.User{ID: 3, Username: "bob", Role: authmodule.RoleAdmin, Status: authmodule.StatusActive},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["total"].(float64)) != 2 {
		t.Fatalf("total = %v, want %d", data["total"], 2)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("items type = %T, want []any", data["items"])
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want %d", len(items), 2)
	}
}

func TestNonAdminCannotAccessUserAdminEndpoints(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleUser, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleUser))
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

func TestCreateUserRejectsDuplicateUsername(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleAdmin, &authmodule.User{ID: 2, Username: "alice", Role: authmodule.RoleUser, Status: authmodule.StatusActive})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", strings.NewReader(`{"username":"alice","password":"secret123","role":"user","status":"active"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "username already exists" {
		t.Fatalf("message = %q, want %q", got.Message, "username already exists")
	}
}

func TestAdminCanUpdateUser(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleAdmin, &authmodule.User{ID: 2, Username: "alice", Role: authmodule.RoleUser, Status: authmodule.StatusActive})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2", strings.NewReader(`{"role":"admin","status":"disabled"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["role"] != authmodule.RoleAdmin {
		t.Fatalf("role = %v, want %q", data["role"], authmodule.RoleAdmin)
	}
	if data["status"] != authmodule.StatusDisabled {
		t.Fatalf("status = %v, want %q", data["status"], authmodule.StatusDisabled)
	}
}

func TestAdminCanDeleteUser(t *testing.T) {
	handler := newAdminUserHandler(t, authmodule.RoleAdmin, &authmodule.User{ID: 2, Username: "alice", Role: authmodule.RoleUser, Status: authmodule.StatusActive})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/2", nil)
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["id"].(float64)) != 2 {
		t.Fatalf("deleted id = %v, want %d", data["id"], 2)
	}
}

func newAdminUserHandler(t *testing.T, actorRole string, users ...*authmodule.User) http.Handler {
	t.Helper()
	service := user.NewService(newUserRepoStub(users...), authmodule.HashPassword)

	apiMux := http.NewServeMux()
	api := humago.New(apiMux, huma.DefaultConfig("Test API", "1.0.0"))
	user.RegisterRoutes(api, service)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/v1/admin/", middleware.RequireAdmin()(middleware.RequireAuthenticated()(apiMux)))
	rootMux.Handle("/", apiMux)
	return middleware.Authenticate(newActorAuthService(t, actorRole))(rootMux)
}

func newActorAuthService(t *testing.T, role string) *authmodule.Service {
	t.Helper()
	actor := &authmodule.User{ID: 1, Username: "operator", Role: role, Status: authmodule.StatusActive}
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	return authmodule.NewService(newAuthRepoStub(actor), tokens)
}

func mustActorToken(t *testing.T, role string) string {
	t.Helper()
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	token, _, err := tokens.Sign(authmodule.CurrentUser{ID: 1, Username: "operator", Role: role, Status: authmodule.StatusActive})
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

func cloneAuthUser(item *authmodule.User) *authmodule.User {
	if item == nil {
		return nil
	}
	clone := *item
	return &clone
}

func mustUint(t *testing.T, value string) uint {
	t.Helper()
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		t.Fatalf("ParseUint(%q) error = %v", value, err)
	}
	return uint(parsed)
}
