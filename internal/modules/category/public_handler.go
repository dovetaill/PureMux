package category

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

func RegisterPublicRoutes(api huma.API, service publicRouteService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "public-category-list", Method: http.MethodGet, Path: "/api/v1/categories", Summary: "list categories"}, func(ctx context.Context, input *categoryListRequest) (*categoryEnvelopeOutput, error) {
		result, err := service.ListPublic(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &categoryEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &categoryEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("category list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})
}
