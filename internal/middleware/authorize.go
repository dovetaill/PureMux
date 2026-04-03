package middleware

import (
	"net/http"

	"github.com/dovetaill/PureMux/internal/identity"
	"github.com/dovetaill/PureMux/internal/modules/auth"
)

func RequireAuthenticated() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := auth.CurrentUserFromContext(r.Context()); !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAdmin() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := identity.PrincipalFromContext(r.Context())
			if ok {
				if principal.Kind != identity.PrincipalAdmin {
					writeAuthError(w, http.StatusForbidden, "forbidden")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			currentUser, ok := auth.CurrentUserFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if currentUser.Role != auth.RoleAdmin {
				writeAuthError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireMember() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := identity.PrincipalFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if principal.Kind != identity.PrincipalMember {
				writeAuthError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
