package category

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

func RegisterAdminRoutes(api huma.API, service adminRouteService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "admin-category-create", Method: http.MethodPost, Path: "/api/v1/admin/categories", Summary: "create category"}, func(ctx context.Context, input *categoryCreateRequest) (*categoryEnvelopeOutput, error) {
		item, err := service.Create(ctx, CreateInput{Name: input.Body.Name, Slug: input.Body.Slug, Description: input.Body.Description})
		if err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusCreated, Body: response.OK("category created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-list", Method: http.MethodGet, Path: "/api/v1/admin/categories", Summary: "list categories"}, func(ctx context.Context, input *categoryListRequest) (*categoryEnvelopeOutput, error) {
		result, err := service.List(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("category list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-get", Method: http.MethodGet, Path: "/api/v1/admin/categories/{id}", Summary: "get category"}, func(ctx context.Context, input *categoryIDRequest) (*categoryEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		item, err := service.Get(ctx, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusOK, Body: response.OK("category detail", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-update", Method: http.MethodPatch, Path: "/api/v1/admin/categories/{id}", Summary: "update category"}, func(ctx context.Context, input *categoryUpdateRequest) (*categoryEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		item, err := service.Update(ctx, UpdateInput{ID: id, Name: stringValue(input.Body.Name), Slug: stringValue(input.Body.Slug), Description: input.Body.Description})
		if err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusOK, Body: response.OK("category updated", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-category-delete", Method: http.MethodDelete, Path: "/api/v1/admin/categories/{id}", Summary: "delete category"}, func(ctx context.Context, input *categoryIDRequest) (*categoryEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid category input")}, nil
		}
		if err := service.Delete(ctx, id); err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusOK, Body: response.OK("category deleted", map[string]uint{"id": id})}, nil
	})
}
