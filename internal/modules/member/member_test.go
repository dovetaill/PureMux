package member_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/identity"
	"github.com/dovetaill/PureMux/internal/middleware"
	membermodule "github.com/dovetaill/PureMux/internal/modules/member"
	"github.com/dovetaill/PureMux/pkg/config"
)

type memberRepoStub struct {
	nextID        uint
	membersByID   map[uint]*membermodule.Member
	membersByName map[string]*membermodule.Member
}

func newMemberRepoStub(items ...*membermodule.Member) *memberRepoStub {
	stub := &memberRepoStub{
		nextID:        1,
		membersByID:   make(map[uint]*membermodule.Member, len(items)),
		membersByName: make(map[string]*membermodule.Member, len(items)),
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		stub.membersByID[item.ID] = &clone
		stub.membersByName[item.Username] = &clone
		if item.ID >= stub.nextID {
			stub.nextID = item.ID + 1
		}
	}
	return stub
}

func (s *memberRepoStub) Create(ctx context.Context, item *membermodule.Member) error {
	if _, ok := s.membersByName[item.Username]; ok {
		return membermodule.ErrDuplicateMember
	}
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	s.membersByID[clone.ID] = &clone
	s.membersByName[clone.Username] = &clone
	*item = clone
	return nil
}

func (s *memberRepoStub) FindByUsername(ctx context.Context, username string) (*membermodule.Member, error) {
	item, ok := s.membersByName[username]
	if !ok {
		return nil, membermodule.ErrMemberNotFound
	}
	clone := *item
	return &clone, nil
}

func (s *memberRepoStub) FindByID(ctx context.Context, id uint) (*membermodule.Member, error) {
	item, ok := s.membersByID[id]
	if !ok {
		return nil, membermodule.ErrMemberNotFound
	}
	clone := *item
	return &clone, nil
}

func TestMemberRegister(t *testing.T) {
	handler, _ := newMemberHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/member/auth/register", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if token, _ := data["token"].(string); token == "" {
		t.Fatal("token = empty, want non-empty")
	}
	memberData, ok := data["member"].(map[string]any)
	if !ok {
		t.Fatalf("member type = %T, want map[string]any", data["member"])
	}
	if memberData["username"] != "alice" {
		t.Fatalf("username = %v, want %q", memberData["username"], "alice")
	}
}

func TestMemberLogin(t *testing.T) {
	passwordHash := mustHashPassword(t, "secret123")
	handler, _ := newMemberHandler(t, &membermodule.Member{
		ID:           1,
		Username:     "alice",
		PasswordHash: passwordHash,
		Status:       membermodule.StatusActive,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/member/auth/login", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if token, _ := data["token"].(string); token == "" {
		t.Fatal("token = empty, want non-empty")
	}
}

func TestMemberCanFetchSelfProfile(t *testing.T) {
	passwordHash := mustHashPassword(t, "secret123")
	handler, service := newMemberHandler(t, &membermodule.Member{
		ID:           7,
		Username:     "alice",
		PasswordHash: passwordHash,
		Status:       membermodule.StatusActive,
	})
	token := mustLogin(t, service, "alice", "secret123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["username"] != "alice" {
		t.Fatalf("username = %v, want %q", data["username"], "alice")
	}
	if data["role"] != membermodule.RoleMember {
		t.Fatalf("role = %v, want %q", data["role"], membermodule.RoleMember)
	}
}

func newMemberHandler(t *testing.T, items ...*membermodule.Member) (http.Handler, *membermodule.Service) {
	t.Helper()
	tokens := identity.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	service := membermodule.NewService(newMemberRepoStub(items...), tokens)

	apiMux := http.NewServeMux()
	api := humago.New(apiMux, huma.DefaultConfig("Test API", "1.0.0"))
	membermodule.RegisterPublicRoutes(api, service)
	membermodule.RegisterSelfRoutes(api, service)

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/v1/me", middleware.RequireMember()(middleware.RequireAuthenticated()(apiMux)))
	rootMux.Handle("/", apiMux)

	return middleware.Authenticate(service)(rootMux), service
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := identity.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	return hash
}

func mustLogin(t *testing.T, svc *membermodule.Service, username, password string) string {
	t.Helper()
	result, err := svc.Login(context.Background(), username, password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	return result.Token
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
