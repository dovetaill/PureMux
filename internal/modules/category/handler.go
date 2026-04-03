package category

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

type routeService interface {
	Create(ctx context.Context, input CreateInput) (*Category, error)
	List(ctx context.Context, page, pageSize int) (*ListResult, error)
	Get(ctx context.Context, id uint) (*Category, error)
	Update(ctx context.Context, input UpdateInput) (*Category, error)
	Delete(ctx context.Context, id uint) error
}

type createInput struct {
	Body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
}

type listInput struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type idInput struct {
	ID string `path:"id"`
}

type updateInput struct {
	ID   string `path:"id"`
	Body struct {
		Name        *string `json:"name,omitempty"`
		Slug        *string `json:"slug,omitempty"`
		Description *string `json:"description,omitempty"`
	}
}

type envelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "admin-category-create", Method: http.MethodPost, Path: "/api/v1/admin/categories", Summary: "create category"}, func(ctx context.Context, input *createInput) (*envelopeOutput, error) {
		item, err := service.Create(ctx, CreateInput{Name: input.Body.Name, Slug: input.Body.Slug, Description: input.Body.Description})
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusCreated, Body: response.OK("category created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-list", Method: http.MethodGet, Path: "/api/v1/admin/categories", Summary: "list categories"}, func(ctx context.Context, input *listInput) (*envelopeOutput, error) {
		result, err := service.List(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("category list", result)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-get", Method: http.MethodGet, Path: "/api/v1/admin/categories/{id}", Summary: "get category"}, func(ctx context.Context, input *idInput) (*envelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		item, err := service.Get(ctx, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("category detail", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-update", Method: http.MethodPatch, Path: "/api/v1/admin/categories/{id}", Summary: "update category"}, func(ctx context.Context, input *updateInput) (*envelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		item, err := service.Update(ctx, UpdateInput{ID: id, Name: stringValue(input.Body.Name), Slug: stringValue(input.Body.Slug), Description: input.Body.Description})
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("category updated", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-delete", Method: http.MethodDelete, Path: "/api/v1/admin/categories/{id}", Summary: "delete category"}, func(ctx context.Context, input *idInput) (*envelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		if err := service.Delete(ctx, id); err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("category deleted", map[string]uint{"id": id})}, nil
	})
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
