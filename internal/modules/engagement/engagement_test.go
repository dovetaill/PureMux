package engagement_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/middleware"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	engagementmodule "github.com/dovetaill/PureMux/internal/modules/engagement"
)

type engagementRepoStub struct {
	nextLikeID     uint
	nextFavoriteID uint
	likes          map[string]*engagementmodule.Like
	favorites      map[string]*engagementmodule.Favorite
}

func newEngagementRepoStub(likes []*engagementmodule.Like, favorites []*engagementmodule.Favorite) *engagementRepoStub {
	stub := &engagementRepoStub{
		nextLikeID:     1,
		nextFavoriteID: 1,
		likes:          make(map[string]*engagementmodule.Like, len(likes)),
		favorites:      make(map[string]*engagementmodule.Favorite, len(favorites)),
	}
	for _, item := range likes {
		if item == nil {
			continue
		}
		clone := *item
		stub.likes[engagementKey(item.MemberID, item.ArticleID)] = &clone
		if item.ID >= stub.nextLikeID {
			stub.nextLikeID = item.ID + 1
		}
	}
	for _, item := range favorites {
		if item == nil {
			continue
		}
		clone := *item
		stub.favorites[engagementKey(item.MemberID, item.ArticleID)] = &clone
		if item.ID >= stub.nextFavoriteID {
			stub.nextFavoriteID = item.ID + 1
		}
	}
	return stub
}

func (s *engagementRepoStub) CreateLike(ctx context.Context, item *engagementmodule.Like) error {
	key := engagementKey(item.MemberID, item.ArticleID)
	if _, ok := s.likes[key]; ok {
		return engagementmodule.ErrDuplicateLike
	}
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextLikeID
		s.nextLikeID++
	}
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = time.Now()
	}
	s.likes[key] = &clone
	*item = clone
	return nil
}

func (s *engagementRepoStub) DeleteLike(ctx context.Context, memberID, articleID uint) error {
	key := engagementKey(memberID, articleID)
	if _, ok := s.likes[key]; !ok {
		return engagementmodule.ErrLikeNotFound
	}
	delete(s.likes, key)
	return nil
}

func (s *engagementRepoStub) CreateFavorite(ctx context.Context, item *engagementmodule.Favorite) error {
	key := engagementKey(item.MemberID, item.ArticleID)
	if _, ok := s.favorites[key]; ok {
		return engagementmodule.ErrDuplicateFavorite
	}
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextFavoriteID
		s.nextFavoriteID++
	}
	if clone.CreatedAt.IsZero() {
		clone.CreatedAt = time.Now()
	}
	s.favorites[key] = &clone
	*item = clone
	return nil
}

func (s *engagementRepoStub) DeleteFavorite(ctx context.Context, memberID, articleID uint) error {
	key := engagementKey(memberID, articleID)
	if _, ok := s.favorites[key]; !ok {
		return engagementmodule.ErrFavoriteNotFound
	}
	delete(s.favorites, key)
	return nil
}

func (s *engagementRepoStub) ListFavorites(ctx context.Context, memberID uint, page, pageSize int) ([]engagementmodule.Favorite, int64, error) {
	items := make([]engagementmodule.Favorite, 0, len(s.favorites))
	for _, item := range s.favorites {
		if item.MemberID != memberID {
			continue
		}
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ArticleID == items[j].ArticleID {
			return items[i].ID < items[j].ID
		}
		return items[i].ArticleID < items[j].ArticleID
	})
	return items, int64(len(items)), nil
}

func TestMemberCanLikeArticle(t *testing.T) {
	handler := newEngagementHandler(t, 7, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles/12/likes", nil)
	req.Header.Set("Authorization", "Bearer member-7")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["member_id"].(float64)) != 7 {
		t.Fatalf("member_id = %v, want %d", data["member_id"], 7)
	}
	if int(data["article_id"].(float64)) != 12 {
		t.Fatalf("article_id = %v, want %d", data["article_id"], 12)
	}
}

func TestMemberCanFavoriteArticle(t *testing.T) {
	handler := newEngagementHandler(t, 7, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles/12/favorites", nil)
	req.Header.Set("Authorization", "Bearer member-7")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["member_id"].(float64)) != 7 {
		t.Fatalf("member_id = %v, want %d", data["member_id"], 7)
	}
	if int(data["article_id"].(float64)) != 12 {
		t.Fatalf("article_id = %v, want %d", data["article_id"], 12)
	}
}

func TestDuplicateFavoriteReturnsConflict(t *testing.T) {
	handler := newEngagementHandler(t, 7, nil, []*engagementmodule.Favorite{{ID: 1, MemberID: 7, ArticleID: 12, CreatedAt: time.Now()}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles/12/favorites", nil)
	req.Header.Set("Authorization", "Bearer member-7")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "favorite already exists" {
		t.Fatalf("message = %q, want %q", got.Message, "favorite already exists")
	}
}

func TestAnonymousCannotFavoriteArticle(t *testing.T) {
	handler := newEngagementHandler(t, 7, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles/12/favorites", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	got := decodeEnvelope(t, rec)
	if got.Message != "unauthorized" {
		t.Fatalf("message = %q, want %q", got.Message, "unauthorized")
	}
}

func TestMyFavoritesReturnsMemberScopedList(t *testing.T) {
	handler := newEngagementHandler(t, 7, nil, []*engagementmodule.Favorite{
		{ID: 1, MemberID: 7, ArticleID: 12, CreatedAt: time.Now()},
		{ID: 2, MemberID: 8, ArticleID: 99, CreatedAt: time.Now()},
		{ID: 3, MemberID: 7, ArticleID: 20, CreatedAt: time.Now()},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/favorites?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer member-7")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if int(data["total"].(float64)) != 2 {
		t.Fatalf("total = %v, want %d", data["total"], 2)
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("items type = %T, want []any", data["items"])
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want %d", len(items), 2)
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("first item type = %T, want map[string]any", items[0])
	}
	if int(first["member_id"].(float64)) != 7 {
		t.Fatalf("member_id = %v, want %d", first["member_id"], 7)
	}
}

func newEngagementHandler(t *testing.T, memberID uint, likes []*engagementmodule.Like, favorites []*engagementmodule.Favorite) http.Handler {
	t.Helper()
	service := engagementmodule.NewService(newEngagementRepoStub(likes, favorites))
	authenticator := &memberAuthenticatorStub{
		token: "member-" + strconv.FormatUint(uint64(memberID), 10),
		user: authmodule.CurrentUser{
			ID:       memberID,
			Username: "member",
			Role:     "member",
			Status:   "active",
		},
	}

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	engagementmodule.RegisterRoutes(api, service)

	return middleware.Authenticate(authenticator)(middleware.RequireMember()(middleware.RequireAuthenticated()(mux)))
}

type memberAuthenticatorStub struct {
	token string
	user  authmodule.CurrentUser
}

func (s *memberAuthenticatorStub) Authenticate(ctx context.Context, token string) (*authmodule.CurrentUser, error) {
	if token != s.token {
		return nil, authmodule.ErrUnauthorized
	}
	user := s.user
	return &user, nil
}

func engagementKey(memberID, articleID uint) string {
	return strconv.FormatUint(uint64(memberID), 10) + ":" + strconv.FormatUint(uint64(articleID), 10)
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
