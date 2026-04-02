package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dovetaill/PureMux/internal/api/register"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/pkg/config"
	"github.com/dovetaill/PureMux/pkg/database"
)

func TestHealthzReturnsAlive(t *testing.T) {
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "PureMux"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Resources: &database.Resources{},
	}

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Message != "alive" {
		t.Fatalf("message = %q, want %q", got.Message, "alive")
	}
}

func TestReadyzReturnsDependencyStatus(t *testing.T) {
	rt := &bootstrap.Runtime{
		Config: &config.Config{
			App:  config.AppConfig{Name: "PureMux"},
			Docs: config.DocsConfig{Enabled: true, OpenAPIPath: "/openapi.json", UIPath: "/docs"},
			HTTP: config.HTTPConfig{ReadTimeoutSeconds: 15},
		},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	handler := register.NewRouter(rt)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	deps, ok := data["dependencies"].(map[string]any)
	if !ok {
		t.Fatalf("dependencies type = %T, want map[string]any", data["dependencies"])
	}
	if deps["database"] != "down" {
		t.Fatalf("dependencies.database = %v, want %q", deps["database"], "down")
	}
	if deps["redis"] != "down" {
		t.Fatalf("dependencies.redis = %v, want %q", deps["redis"], "down")
	}
}

func TestResponseHelpersReturnStandardShape(t *testing.T) {
	ok := response.OK("ok", map[string]any{"name": "puremux"})
	if ok.Code != 0 {
		t.Fatalf("OK code = %d, want %d", ok.Code, 0)
	}
	if ok.Message != "ok" {
		t.Fatalf("OK message = %q, want %q", ok.Message, "ok")
	}
	if ok.Data == nil {
		t.Fatal("OK data = nil, want non-nil")
	}

	fail := response.Fail(1001, "bad request")
	if fail.Code != 1001 {
		t.Fatalf("Fail code = %d, want %d", fail.Code, 1001)
	}
	if fail.Message != "bad request" {
		t.Fatalf("Fail message = %q, want %q", fail.Message, "bad request")
	}
}
