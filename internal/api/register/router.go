package register

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/handlers"
	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/internal/identity"
	"github.com/dovetaill/PureMux/internal/middleware"
	articlemodule "github.com/dovetaill/PureMux/internal/modules/article"
	"github.com/dovetaill/PureMux/internal/modules/auth"
	categorymodule "github.com/dovetaill/PureMux/internal/modules/category"
	engagementmodule "github.com/dovetaill/PureMux/internal/modules/engagement"
	membermodule "github.com/dovetaill/PureMux/internal/modules/member"
	usermodule "github.com/dovetaill/PureMux/internal/modules/user"
	"github.com/dovetaill/PureMux/pkg/config"
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
	publicRoutes := huma.NewGroup(api)
	memberAuthRoutes := huma.NewGroup(api)
	memberSelfRoutes := huma.NewGroup(api)
	adminRoutes := huma.NewGroup(api)
	memberSelfRoutes.UseMiddleware(requireMemberMiddleware(api))
	adminRoutes.UseMiddleware(requireAdminMiddleware(api))

	handlers.RegisterHealth(publicRoutes)
	handlers.RegisterReady(publicRoutes, rt)

	handler := http.Handler(apiMux)
	authenticators := make([]tokenAuthenticator, 0, 2)
	if authService := newAuthService(rt); authService != nil {
		authenticators = append(authenticators, authService)
		auth.RegisterRoutes(publicRoutes, authService)
		if userService := newUserService(rt); userService != nil {
			usermodule.RegisterRoutes(adminRoutes, userService)
		}
		if categoryService := newCategoryService(rt); categoryService != nil {
			categorymodule.RegisterPublicRoutes(publicRoutes, categoryService)
			categorymodule.RegisterAdminRoutes(adminRoutes, categoryService)
		}
		if articleService := newArticleService(rt); articleService != nil {
			articlemodule.RegisterPublicRoutes(publicRoutes, articleService)
			articlemodule.RegisterAdminRoutes(adminRoutes, articleService)
		}
	}
	if memberService := newMemberService(rt); memberService != nil {
		authenticators = append(authenticators, memberService)
		membermodule.RegisterPublicRoutes(memberAuthRoutes, memberService)
		membermodule.RegisterSelfRoutes(memberSelfRoutes, memberService)
	}
	if engagementService := newEngagementService(rt); engagementService != nil {
		engagementmodule.RegisterRoutes(memberSelfRoutes, engagementService)
	}
	if len(authenticators) > 0 {
		handler = middleware.Authenticate(compositeAuthenticator{authenticators: authenticators})(apiMux)
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

func newUserService(rt *bootstrap.Runtime) *usermodule.Service {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := usermodule.NewRepository(rt.Resources.MySQL)
	return usermodule.NewService(repo, auth.HashPassword)
}

func newCategoryService(rt *bootstrap.Runtime) *categorymodule.Service {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := categorymodule.NewRepository(rt.Resources.MySQL)
	return categorymodule.NewService(repo)
}

func newArticleService(rt *bootstrap.Runtime) *articlemodule.Service {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := articlemodule.NewRepository(rt.Resources.MySQL)
	return articlemodule.NewService(repo)
}

func newMemberService(rt *bootstrap.Runtime) *membermodule.Service {
	if rt == nil || rt.Config == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := membermodule.NewRepository(rt.Resources.MySQL)
	return membermodule.NewService(repo, identity.NewTokenManager(memberJWTConfig(rt.Config.Auth.JWT)))
}

func newEngagementService(rt *bootstrap.Runtime) *engagementmodule.Service {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}
	repo := engagementmodule.NewRepository(rt.Resources.MySQL)
	return engagementmodule.NewService(repo)
}

func nilLogger(rt *bootstrap.Runtime) *slog.Logger {
	if rt == nil {
		return nil
	}
	return rt.Logger
}

func requireAdminMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if principal, ok := identity.PrincipalFromContext(ctx.Context()); ok {
			if principal.Kind != identity.PrincipalAdmin {
				_ = huma.WriteErr(api, ctx, http.StatusForbidden, "forbidden")
				return
			}
			next(ctx)
			return
		}

		currentUser, ok := auth.CurrentUserFromContext(ctx.Context())
		if !ok {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		if currentUser.Role != auth.RoleAdmin {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "forbidden")
			return
		}
		next(ctx)
	}
}

func requireMemberMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		principal, ok := identity.PrincipalFromContext(ctx.Context())
		if !ok {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized")
			return
		}
		if principal.Kind != identity.PrincipalMember {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "forbidden")
			return
		}
		next(ctx)
	}
}

type tokenAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*identity.Actor, error)
}

type compositeAuthenticator struct {
	authenticators []tokenAuthenticator
}

func (c compositeAuthenticator) Authenticate(ctx context.Context, token string) (*identity.Actor, error) {
	for _, authenticator := range c.authenticators {
		if authenticator == nil {
			continue
		}
		actor, err := authenticator.Authenticate(ctx, token)
		if err == nil {
			return actor, nil
		}
		if errors.Is(err, identity.ErrUnauthorized) {
			continue
		}
		return nil, err
	}
	return nil, identity.ErrUnauthorized
}

func memberJWTConfig(base config.JWTConfig) config.JWTConfig {
	cfg := base
	cfg.Secret = strings.TrimSpace(cfg.Secret) + ":member"
	cfg.Issuer = strings.TrimSpace(cfg.Issuer) + "/member"
	if strings.TrimSpace(base.Secret) == "" {
		cfg.Secret = "PureMux:member"
	}
	if strings.TrimSpace(base.Issuer) == "" {
		cfg.Issuer = "PureMux/member"
	}
	return cfg
}
