package category

import (
	"context"
	"strconv"

	"github.com/dovetaill/PureMux/internal/api/response"
)

type publicRouteService interface {
	ListPublic(ctx context.Context, page, pageSize int) (*ListResult, error)
}

type adminRouteService interface {
	Create(ctx context.Context, input CreateInput) (*Category, error)
	List(ctx context.Context, page, pageSize int) (*ListResult, error)
	Get(ctx context.Context, id uint) (*Category, error)
	Update(ctx context.Context, input UpdateInput) (*Category, error)
	Delete(ctx context.Context, id uint) error
}

type categoryCreateBody struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type categoryCreateRequest struct {
	Body categoryCreateBody
}

type categoryListRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type categoryIDRequest struct {
	ID string `path:"id"`
}

type categoryUpdateBody struct {
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
}

type categoryUpdateRequest struct {
	ID   string `path:"id"`
	Body categoryUpdateBody
}

type categoryEnvelopeOutput struct {
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
