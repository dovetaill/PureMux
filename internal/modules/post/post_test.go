package post_test

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
	postmodule "github.com/dovetaill/PureMux/internal/modules/post"
)

type postRepoStub struct {
	nextID uint
	items  map[uint]*postmodule.Post
}

func newPostRepoStub(items ...*postmodule.Post) *postRepoStub {
	stub := &postRepoStub{nextID: 1, items: make(map[uint]*postmodule.Post, len(items))}
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

func (s *postRepoStub) Create(ctx context.Context, item *postmodule.Post) error {
	for _, existing := range s.items {
		if existing.Slug == item.Slug {
			return postmodule.ErrDuplicatePostSlug
		}
	}
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	s.items[clone.ID] = &clone
	*item = clone
	return nil
}

func (s *postRepoStub) List(ctx context.Context, page, pageSize int) ([]postmodule.Post, int64, error) {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	items := make([]postmodule.Post, 0, len(ids))
	for _, id := range ids {
		items = append(items, *s.items[uint(id)])
	}
	return items, int64(len(items)), nil
}

func (s *postRepoStub) FindByID(ctx context.Context, id uint) (*postmodule.Post, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, postmodule.ErrPostNotFound
	}
	clone := *item
	return &clone, nil
}

func (s *postRepoStub) FindBySlug(ctx context.Context, slug string) (*postmodule.Post, error) {
	for _, item := range s.items {
		if item.Slug != slug {
			continue
		}
		clone := *item
		return &clone, nil
	}
	return nil, postmodule.ErrPostNotFound
}

func (s *postRepoStub) Update(ctx context.Context, item *postmodule.Post) error {
	if _, ok := s.items[item.ID]; !ok {
		return postmodule.ErrPostNotFound
	}
	for _, existing := range s.items {
		if existing.ID != item.ID && existing.Slug == item.Slug {
			return postmodule.ErrDuplicatePostSlug
		}
	}
	clone := *item
	clone.UpdatedAt = time.Now()
	s.items[item.ID] = &clone
	*item = clone
	return nil
}

func (s *postRepoStub) Delete(ctx context.Context, id uint) error {
	if _, ok := s.items[id]; !ok {
		return postmodule.ErrPostNotFound
	}
	delete(s.items, id)
	return nil
}

func TestCreatePost(t *testing.T) {
	handler, _ := newPostHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"First Post","summary":"intro","content":"hello world"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["title"] != "First Post" {
		t.Fatalf("title = %v, want %q", data["title"], "First Post")
	}
	if data["slug"] != "first-post" {
		t.Fatalf("slug = %v, want %q", data["slug"], "first-post")
	}
}

func TestListPosts(t *testing.T) {
	handler, _ := newPostHandler(t,
		&postmodule.Post{ID: 1, Title: "First", Slug: "first", Summary: "intro", Content: "hello"},
		&postmodule.Post{ID: 2, Title: "Second", Slug: "second", Summary: "more", Content: "world"},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?page=1&page_size=20", nil)
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
}

func TestGetPost(t *testing.T) {
	handler, _ := newPostHandler(t, &postmodule.Post{ID: 7, Title: "Seven", Slug: "seven", Summary: "lucky", Content: "content"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/7", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["id"] != float64(7) {
		t.Fatalf("id = %v, want %d", data["id"], 7)
	}
}

func TestUpdatePost(t *testing.T) {
	handler, _ := newPostHandler(t, &postmodule.Post{ID: 3, Title: "Draft", Slug: "draft", Summary: "intro", Content: "hello"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/posts/3", strings.NewReader(`{"title":"Draft Updated","summary":"changed","content":"updated body"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeEnvelope(t, rec)
	data := envelopeData(t, got)
	if data["title"] != "Draft Updated" {
		t.Fatalf("title = %v, want %q", data["title"], "Draft Updated")
	}
}

func TestDeletePost(t *testing.T) {
	handler, repo := newPostHandler(t, &postmodule.Post{ID: 4, Title: "Remove", Slug: "remove", Summary: "delete", Content: "bye"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/4", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(repo.items) != 0 {
		t.Fatalf("remaining items = %d, want %d", len(repo.items), 0)
	}
}

func newPostHandler(t *testing.T, items ...*postmodule.Post) (http.Handler, *postRepoStub) {
	t.Helper()
	repo := newPostRepoStub(items...)
	service := postmodule.NewService(repo)

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	postmodule.RegisterRoutes(api, service)

	return mux, repo
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
