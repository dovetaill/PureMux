package register_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dovetaill/PureMux/internal/api/register"
	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/internal/identity"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	"github.com/dovetaill/PureMux/pkg/config"
	"github.com/dovetaill/PureMux/pkg/database"
	"gorm.io/gorm"
)

type openAPIDocument struct {
	Paths map[string]map[string]any `json:"paths"`
}

func TestRouterRegistersAuthAndBusinessRoutes(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime())

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var doc openAPIDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}

	assertOperation(t, doc.Paths, "/api/v1/auth/login", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/auth/me", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/member/auth/register", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/member/auth/login", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/me", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/me/favorites", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/users", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/admin/users", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/users/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/users/{id}", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/admin/users/{id}", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/admin/categories", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/admin/categories", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/categories/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/categories/{id}", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/admin/categories/{id}", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/categories", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/articles", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/articles/{slug}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/articles/{id}/likes", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/articles/{id}/likes", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/articles/{id}/favorites", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/articles/{id}/favorites", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles/{id}", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles/{id}", http.MethodDelete)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles/{id}/publish", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/admin/articles/{id}/unpublish", http.MethodPost)
}

func TestPublicArticleRoutesAreAccessibleWithoutAuth(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles?page=1&page_size=20", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("status = %d, want non-%d for unauthenticated public article route", rec.Code, http.StatusUnauthorized)
	}
}

func TestMemberRoutesRequireMemberAuth(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d for unauthenticated member route", rec.Code, http.StatusUnauthorized)
	}
}

func TestAdminRoutesRejectNonAdminPrincipal(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime())
	currentUser := authmodule.CurrentUser{
		ID:       7,
		Username: "writer",
		Role:     authmodule.RoleUser,
		Status:   authmodule.StatusActive,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req = req.WithContext(authmodule.ContextWithCurrentUser(req.Context(), currentUser))
	req = req.WithContext(identity.ContextWithPrincipal(req.Context(), identity.PrincipalFromActor(currentUser.ToActor())))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d for non-admin principal on admin route", rec.Code, http.StatusForbidden)
	}
}

func newRouterTestRuntime() *bootstrap.Runtime {
	return &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "PureMux"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
			Auth: config.AuthConfig{
				JWT: config.JWTConfig{
					Secret:     "test-secret",
					Issuer:     "puremux-test",
					TTLMinutes: 120,
				},
			},
		},
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{MySQL: &gorm.DB{}},
	}
}

func assertOperation(t *testing.T, paths map[string]map[string]any, path string, method string) {
	t.Helper()

	operations, ok := paths[path]
	if !ok {
		t.Fatalf("missing path %s", path)
	}
	if _, ok := operations[httpMethodKey(method)]; !ok {
		t.Fatalf("missing %s %s operation", method, path)
	}
}

func httpMethodKey(method string) string {
	switch method {
	case http.MethodGet:
		return "get"
	case http.MethodPost:
		return "post"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}
