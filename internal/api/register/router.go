package register

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/handlers"
	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/internal/middleware"
	"github.com/dovetaill/PureMux/internal/modules/auth"
)

// NewRouter 构建基于 Huma 的 HTTP 路由。
func NewRouter(rt *bootstrap.Runtime) http.Handler {
	apiMux := http.NewServeMux()
	cfg := huma.DefaultConfig("PureMux API", "0.1.0")
	if rt != nil && rt.Config != nil {
		if rt.Config.App.Name != "" {
			cfg.Info.Title = rt.Config.App.Name
		}
		cfg.OpenAPIPath = normalizeOpenAPIPath(rt.Config.Docs.OpenAPIPath)
		if rt.Config.Docs.UIPath != "" {
			cfg.DocsPath = rt.Config.Docs.UIPath
		}
	}

	api := humago.New(apiMux, cfg)
	handlers.RegisterHealth(api)
	handlers.RegisterReady(api, rt)

	handler := http.Handler(apiMux)
	if authService := newAuthService(rt); authService != nil {
		auth.RegisterRoutes(api, authService)

		rootMux := http.NewServeMux()
		rootMux.Handle("/api/v1/auth/me", middleware.RequireAuthenticated()(apiMux))
		rootMux.Handle("/api/v1/admin/", middleware.RequireAdmin()(middleware.RequireAuthenticated()(apiMux)))
		rootMux.Handle("/", apiMux)
		handler = middleware.Authenticate(authService)(rootMux)
	}

	timeout := 15 * time.Second
	if rt != nil && rt.Config != nil && rt.Config.HTTP.ReadTimeoutSeconds > 0 {
		timeout = time.Duration(rt.Config.HTTP.ReadTimeoutSeconds) * time.Second
	}

	return middleware.Chain(
		handler,
		middleware.RequestID(),
		middleware.Recover(),
		middleware.Timeout(timeout),
		middleware.AccessLog(nilLogger(rt)),
	)
}

func normalizeOpenAPIPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/openapi"
	}
	if strings.HasSuffix(path, ".json") {
		return strings.TrimSuffix(path, ".json")
	}
	return path
}

func newAuthService(rt *bootstrap.Runtime) *auth.Service {
	if rt == nil || rt.Config == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := auth.NewRepository(rt.Resources.MySQL)
	tokens := auth.NewTokenManager(rt.Config.Auth.JWT)
	return auth.NewService(repo, tokens)
}

func nilLogger(rt *bootstrap.Runtime) *slog.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}
