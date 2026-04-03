package article_test

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
	articlemodule "github.com/dovetaill/PureMux/internal/modules/article"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
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

type articleRepoStub struct {
	nextID uint
	items  map[uint]*articlemodule.Article
}

func newArticleRepoStub(items ...*articlemodule.Article) *articleRepoStub {
	stub := &articleRepoStub{nextID: 1, items: make(map[uint]*articlemodule.Article, len(items))}
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

func (s *articleRepoStub) Create(ctx context.Context, item *articlemodule.Article) error {
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	if clone.Status == "" {
		clone.Status = articlemodule.StatusDraft
	}
	s.items[clone.ID] = &clone
	*item = clone
	return nil
}

func (s *articleRepoStub) List(ctx context.Context, filter articlemodule.ListFilter, page, pageSize int) ([]articlemodule.Article, int64, error) {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	items := make([]articlemodule.Article, 0, len(ids))
	for _, id := range ids {
		item := s.items[uint(id)]
		if filter.AuthorID != nil && item.AuthorID != *filter.AuthorID {
			continue
		}
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (s *articleRepoStub) FindByID(ctx context.Context, id uint) (*articlemodule.Article, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, articlemodule.ErrArticleNotFound
	}
	clone := *item
	return &clone, nil
}

func (s *articleRepoStub) Update(ctx context.Context, item *articlemodule.Article) error {
	if _, ok := s.items[item.ID]; !ok {
		return articlemodule.ErrArticleNotFound
	}
	clone := *item
	clone.UpdatedAt = time.Now()
	s.items[item.ID] = &clone
	*item = clone
	return nil
}

func (s *articleRepoStub) Delete(ctx context.Context, id uint) error {
	if _, ok := s.items[id]; !ok {
		return articlemodule.ErrArticleNotFound
	}
	delete(s.items, id)
	return nil
}

func TestUserCanCreateOwnArticle(t *testing.T) {
	handler := newArticleHandler(t, 7, authmodule.RoleUser)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", strings.NewReader(`{"title":"First Post","summary":"intro","content":"hello","category_id":2}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, 7, authmodule.RoleUser))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["author_id"] != float64(7) {
		t.Fatalf("author_id = %v, want %d", data["author_id"], 7)
	}
	if data["status"] != articlemodule.StatusDraft {
		t.Fatalf("status = %v, want %q", data["status"], articlemodule.StatusDraft)
	}
}

func TestUserCanOnlyListOwnArticles(t *testing.T) {
	handler := newArticleHandler(t, 7, authmodule.RoleUser,
		&articlemodule.Article{ID: 1, Title: "Mine", Status: articlemodule.StatusDraft, AuthorID: 7, CategoryID: 2},
		&articlemodule.Article{ID: 2, Title: "Others", Status: articlemodule.StatusDraft, AuthorID: 8, CategoryID: 3},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, 7, authmodule.RoleUser))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["total"].(float64)) != 1 {
		t.Fatalf("total = %v, want %d", data["total"], 1)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("items type = %T, want []any", data["items"])
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want %d", len(items), 1)
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("first item type = %T, want map[string]any", items[0])
	}
	if first["author_id"] != float64(7) {
		t.Fatalf("author_id = %v, want %d", first["author_id"], 7)
	}
}

func TestUserCannotUpdateOtherUsersArticle(t *testing.T) {
	handler := newArticleHandler(t, 7, authmodule.RoleUser,
		&articlemodule.Article{ID: 1, Title: "Others", Status: articlemodule.StatusDraft, AuthorID: 8, CategoryID: 3},
	)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/articles/1", strings.NewReader(`{"title":"takeover"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, 7, authmodule.RoleUser))
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

func TestAdminCanManageAnyArticle(t *testing.T) {
	handler := newArticleHandler(t, 1, authmodule.RoleAdmin,
		&articlemodule.Article{ID: 1, Title: "Others", Content: "body", Status: articlemodule.StatusDraft, AuthorID: 8, CategoryID: 3},
	)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/articles/1", strings.NewReader(`{"title":"editorial review"}`))
	req.Header.Set("Authorization", "Bearer "+mustActorToken(t, 1, authmodule.RoleAdmin))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["title"] != "editorial review" {
		t.Fatalf("title = %v, want %q", data["title"], "editorial review")
	}
}

func TestPublishAndUnpublishTransitionsStatus(t *testing.T) {
	handler := newArticleHandler(t, 7, authmodule.RoleUser,
		&articlemodule.Article{ID: 1, Title: "Mine", Status: articlemodule.StatusDraft, AuthorID: 7, CategoryID: 2},
	)

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/articles/1/publish", nil)
	publishReq.Header.Set("Authorization", "Bearer "+mustActorToken(t, 7, authmodule.RoleUser))
	publishRec := httptest.NewRecorder()
	handler.ServeHTTP(publishRec, publishReq)

	if publishRec.Code != http.StatusOK {
		t.Fatalf("publish status = %d, want %d", publishRec.Code, http.StatusOK)
	}
	published := decodeEnvelope(t, publishRec)
	publishedData := envelopeData(t, published)
	if publishedData["status"] != articlemodule.StatusPublished {
		t.Fatalf("published status = %v, want %q", publishedData["status"], articlemodule.StatusPublished)
	}

	unpublishReq := httptest.NewRequest(http.MethodPost, "/api/v1/articles/1/unpublish", nil)
	unpublishReq.Header.Set("Authorization", "Bearer "+mustActorToken(t, 7, authmodule.RoleUser))
	unpublishRec := httptest.NewRecorder()
	handler.ServeHTTP(unpublishRec, unpublishReq)

	if unpublishRec.Code != http.StatusOK {
		t.Fatalf("unpublish status = %d, want %d", unpublishRec.Code, http.StatusOK)
	}
	unpublished := decodeEnvelope(t, unpublishRec)
	unpublishedData := envelopeData(t, unpublished)
	if unpublishedData["status"] != articlemodule.StatusDraft {
		t.Fatalf("unpublished status = %v, want %q", unpublishedData["status"], articlemodule.StatusDraft)
	}
}

func newArticleHandler(t *testing.T, actorID uint, actorRole string, items ...*articlemodule.Article) http.Handler {
	t.Helper()
	repo := newArticleRepoStub(items...)
	service := articlemodule.NewService(repo)

	apiMux := http.NewServeMux()
	api := humago.New(apiMux, huma.DefaultConfig("Test API", "1.0.0"))
	articlemodule.RegisterRoutes(api, service)

	return middleware.Authenticate(newActorAuthService(t, actorID, actorRole))(apiMux)
}

func newActorAuthService(t *testing.T, actorID uint, role string) *authmodule.Service {
	t.Helper()
	actor := &authmodule.User{ID: actorID, Username: "operator", Role: role, Status: authmodule.StatusActive}
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	return authmodule.NewService(newAuthRepoStub(actor), tokens)
}

func mustActorToken(t *testing.T, actorID uint, role string) string {
	t.Helper()
	tokens := authmodule.NewTokenManager(config.JWTConfig{Secret: "test-secret", Issuer: "PureMuxTest", TTLMinutes: 60})
	token, _, err := tokens.Sign(authmodule.CurrentUser{ID: actorID, Username: "operator", Role: role, Status: authmodule.StatusActive})
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
