package category_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/middleware"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	categorymodule "github.com/dovetaill/PureMux/internal/modules/category"
	"github.com/dovetaill/PureMux/pkg/config"
)

type authRepoStub struct {
	usersByID map[uint]*authmodule.User
}

func newAuthRepoStub(users ...*authmodule.User) *authRepoStub {
	stub := &authRepoStub{usersByID: make(map[uint]*authmodule.User, len(users))}
	for _, item := range users {
		if item == nil {
			continue
		}
		clone := *item
		stub.usersByID[item.ID] = &clone
	}
	return stub
}

func (s *authRepoStub) FindByUsername(ctx context.Context, username string) (*authmodule.User, error) {
	for _, item := range s.usersByID {
		if item.Username == username {
			clone := *item
			return &clone, nil
		}
	}
	return nil, authmodule.ErrUserNotFound
}

func (s *authRepoStub) FindByID(ctx context.Context, id uint) (*authmodule.User, error) {
	item, ok := s.usersByID[id]
	if !ok {
		return nil, authmodule.ErrUserNotFound
	}
	clone := *item
	return &clone, nil
}

type categoryRepoStub struct {
	nextID uint
	items  map[uint]*categorymodule.Category
}

func newCategoryRepoStub(items ...*categorymodule.Category) *categoryRepoStub {
	stub := &categoryRepoStub{nextID: 1, items: make(map[uint]*categorymodule.Category, len(items))}
	for _, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		stub.items[item.ID] = &clone
		if item.ID >= stub.nextID {
			stub.nextID = item.ID + 1
		}
	}
	return stub
}

func (s *categoryRepoStub) Create(ctx context.Context, item *categorymodule.Category) error {
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	s.items[clone.ID] = &clone
	*item = clone
	return nil
}

func (s *categoryRepoStub) List(ctx context.Context, page, pageSize int) ([]categorymodule.Category, int64, error) {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	items := make([]categorymodule.Category, 0, len(ids))
	for _, id := range ids {
		items = append(items, *s.items[uint(id)])
	}
	return items, int64(len(items)), nil
}

func (s *categoryRepoStub) FindByID(ctx context.Context, id uint) (*categorymodule.Category, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, categorymodule.ErrCategoryNotFound
	}
	clone := *item
	return &clone, nil
}

func (s *categoryRepoStub) FindBySlug(ctx context.Context, slug string) (*categorymodule.Category, error) {
	for _, item := range s.items {
		if item.Slug == slug {
			clone := *item
			return &clone, nil
		}
	}
	return nil, categorymodule.ErrCategoryNotFound
}

func (s *categoryRepoStub) Update(ctx context.Context, item *categorymodule.Category) error {
	if _, ok := s.items[item.ID]; !ok {
		return categorymodule.ErrCategoryNotFound
	}
	clone := *item
	clone.UpdatedAt = time.Now()
	s.items[item.ID] = &clone
	*item = clone
	return nil
}

func (s *categoryRepoStub) Delete(ctx context.Context, id uint) error {
	if _, ok := s.items[id]; !ok {
		return categorymodule.ErrCategoryNotFound
	}
	delete(s.items, id)
	return nil
}

func TestPublicCategoryListIsAccessibleWithoutAuth(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin,
		&categorymodule.Category{ID: 1, Name: "News", Slug: "news", Description: "daily"},
		&categorymodule.Category{ID: 2, Name: "Tech", Slug: "tech", Description: "technology"},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?page=1&page_size=20", nil)
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
}

func TestAdminCategoryCrudStillRequiresAdmin(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleUser)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(`{"name":"News","slug":"news","description":"daily news"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleUser))
	req.Header.Set("Content-Type", "application/json")
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

func TestAdminCanCreateCategory(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(`{"name":"News","slug":"news","description":"daily news"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["name"] != "News" {
		t.Fatalf("name = %v, want %q", data["name"], "News")
	}
	if data["slug"] != "news" {
		t.Fatalf("slug = %v, want %q", data["slug"], "news")
	}
}

func TestAdminCanListCategories(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin,
		&categorymodule.Category{ID: 1, Name: "News", Slug: "news", Description: "daily"},
		&categorymodule.Category{ID: 2, Name: "Tech", Slug: "tech", Description: "technology"},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories?page=1&page_size=20", nil)
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

func TestNonAdminCannotAccessCategoryAdminEndpoints(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleUser)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
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

func TestCreateCategoryRejectsDuplicateSlug(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin, &categorymodule.Category{ID: 1, Name: "News", Slug: "news", Description: "daily"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", strings.NewReader(`{"name":"Duplicate","slug":"news","description":"dup"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "slug already exists" {
		t.Fatalf("message = %q, want %q", got.Message, "slug already exists")
	}
}

func TestAdminCanUpdateCategory(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin, &categorymodule.Category{ID: 1, Name: "News", Slug: "news", Description: "daily"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/categories/1", strings.NewReader(`{"name":"Politics","slug":"politics","description":"updates"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["slug"] != "politics" {
		t.Fatalf("slug = %v, want %q", data["slug"], "politics")
	}
}

func TestAdminCanDeleteCategory(t *testing.T) {
	handler := newAdminCategoryHandler(t, authmodule.RoleAdmin, &categorymodule.Category{ID: 1, Name: "News", Slug: "news", Description: "daily"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/categories/1", nil)
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, authmodule.RoleAdmin))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["id"].(float64)) != 1 {
		t.Fatalf("deleted id = %v, want %d", data["id"], 1)
	}
}

func newAdminCategoryHandler(t *testing.T, actorRole string, items ...*categorymodule.Category) http.Handler {
	t.Helper()
	service := categorymodule.NewService(newCategoryRepoStub(items...))

	apiMux := http.NewServeMux()
	api := humago.New(apiMux, huma.DefaultConfig("Test API", "1.0.0"))
	categorymodule.RegisterPublicRoutes(api, service)
	categorymodule.RegisterAdminRoutes(api, service)

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
