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
)

// NewRouter 构建基于 Huma 的 HTTP 路由。
func NewRouter(rt *bootstrap.Runtime) http.Handler {
	mux := http.NewServeMux()
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

	api := humago.New(mux, cfg)
	handlers.RegisterHealth(api)
	handlers.RegisterReady(api, rt)

	timeout := 15 * time.Second
	if rt != nil && rt.Config != nil && rt.Config.HTTP.ReadTimeoutSeconds > 0 {
		timeout = time.Duration(rt.Config.HTTP.ReadTimeoutSeconds) * time.Second
	}

	return middleware.Chain(
		mux,
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

func nilLogger(rt *bootstrap.Runtime) *slog.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}
