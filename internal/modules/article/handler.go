package article

import (
	"context"
	"strconv"

	"github.com/dovetaill/PureMux/internal/api/response"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

type publicRouteService interface {
	ListPublic(ctx context.Context, page, pageSize int) (*ListResult, error)
	GetPublicBySlug(ctx context.Context, slug string) (*Article, error)
}

type adminRouteService interface {
	Create(ctx context.Context, actor authmodule.CurrentUser, input CreateInput) (*Article, error)
	List(ctx context.Context, actor authmodule.CurrentUser, page, pageSize int) (*ListResult, error)
	Get(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error)
	Update(ctx context.Context, actor authmodule.CurrentUser, input UpdateInput) (*Article, error)
	Delete(ctx context.Context, actor authmodule.CurrentUser, id uint) error
	Publish(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error)
	Unpublish(ctx context.Context, actor authmodule.CurrentUser, id uint) (*Article, error)
}

type articleCreateBody struct {
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Content    string `json:"content"`
	CategoryID uint   `json:"category_id"`
}

type articleCreateRequest struct {
	Body articleCreateBody
}

type articleListRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type articleIDRequest struct {
	ID string `path:"id"`
}

type articleSlugRequest struct {
	Slug string `path:"slug"`
}

type articleUpdateBody struct {
	Title      *string `json:"title,omitempty"`
	Summary    *string `json:"summary,omitempty"`
	Content    *string `json:"content,omitempty"`
	CategoryID *uint   `json:"category_id,omitempty"`
}

type articleUpdateRequest struct {
	ID   string `path:"id"`
	Body articleUpdateBody
}

type articleEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func parseID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
