package article

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

	huma.Register(api, huma.Operation{OperationID: "public-article-list", Method: http.MethodGet, Path: "/api/v1/articles", Summary: "list published articles"}, func(ctx context.Context, input *articleListRequest) (*articleEnvelopeOutput, error) {
		result, err := service.ListPublic(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("article list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "public-article-detail", Method: http.MethodGet, Path: "/api/v1/articles/{slug}", Summary: "get published article by slug"}, func(ctx context.Context, input *articleSlugRequest) (*articleEnvelopeOutput, error) {
		item, err := service.GetPublicBySlug(ctx, input.Slug)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article detail", item)}, nil
	})
}
