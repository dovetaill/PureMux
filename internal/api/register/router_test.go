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
	"github.com/dovetaill/PureMux/pkg/config"
	"github.com/dovetaill/PureMux/pkg/database"
	"gorm.io/gorm"
)

type openAPIDocument struct {
	Paths map[string]map[string]any `json:"paths"`
}

func TestRouterRegistersStarterRoutes(t *testing.T) {
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

	assertOperation(t, doc.Paths, "/healthz", http.MethodGet)
	assertOperation(t, doc.Paths, "/readyz", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts", http.MethodPost)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodGet)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodPatch)
	assertOperation(t, doc.Paths, "/api/v1/posts/{id}", http.MethodDelete)

	assertPathAbsent(t, doc.Paths, "/api/v1/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/auth/me")
	assertPathAbsent(t, doc.Paths, "/api/v1/member/auth/register")
	assertPathAbsent(t, doc.Paths, "/api/v1/member/auth/login")
	assertPathAbsent(t, doc.Paths, "/api/v1/me")
	assertPathAbsent(t, doc.Paths, "/api/v1/me/favorites")
	assertPathAbsent(t, doc.Paths, "/api/v1/admin/users")
	assertPathAbsent(t, doc.Paths, "/api/v1/admin/categories")
	assertPathAbsent(t, doc.Paths, "/api/v1/categories")
	assertPathAbsent(t, doc.Paths, "/api/v1/articles")
	assertPathAbsent(t, doc.Paths, "/api/v1/admin/articles")
}

func TestRouterServesStarterHealthAndDocsEndpoints(t *testing.T) {
	handler := register.NewRouter(newRouterTestRuntime())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code >= http.StatusBadRequest {
		t.Fatalf("docs status = %d, want < %d", rec.Code, http.StatusBadRequest)
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

func assertPathAbsent(t *testing.T, paths map[string]map[string]any, path string) {
	t.Helper()
	if _, ok := paths[path]; ok {
		t.Fatalf("path %s should not be registered on starter router", path)
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
