package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dovetaill/PureMux/internal/identity"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

type stubAuthenticator struct {
	user *authmodule.CurrentUser
	err  error
}

func (s stubAuthenticator) Authenticate(ctx context.Context, token string) (*authmodule.CurrentUser, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestAuthenticateStoresAdminPrincipal(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		user: &authmodule.CurrentUser{
			ID:       1,
			Username: "admin",
			Role:     authmodule.RoleAdmin,
			Status:   authmodule.StatusActive,
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := identity.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		if principal.Kind != identity.PrincipalAdmin {
			t.Fatalf("principal kind = %q, want %q", principal.Kind, identity.PrincipalAdmin)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAuthenticateStoresMemberPrincipal(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		user: &authmodule.CurrentUser{
			ID:       9,
			Username: "reader",
			Role:     "member",
			Status:   authmodule.StatusActive,
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := identity.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		if principal.Kind != identity.PrincipalMember {
			t.Fatalf("principal kind = %q, want %q", principal.Kind, identity.PrincipalMember)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAdminRejectsMemberPrincipal(t *testing.T) {
	handler := Authenticate(stubAuthenticator{
		user: &authmodule.CurrentUser{
			ID:       11,
			Username: "member",
			Role:     "member",
			Status:   authmodule.StatusActive,
		},
	})(RequireAdmin()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireMemberRejectsAnonymous(t *testing.T) {
	handler := RequireMember()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
