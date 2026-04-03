package article

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

func RegisterAdminRoutes(api huma.API, service adminRouteService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "admin-article-create", Method: http.MethodPost, Path: "/api/v1/admin/articles", Summary: "create article"}, func(ctx context.Context, input *articleCreateRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		item, err := service.Create(ctx, actor, CreateInput{Title: input.Body.Title, Summary: input.Body.Summary, Content: input.Body.Content, CategoryID: input.Body.CategoryID})
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusCreated, Body: response.OK("article created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-list", Method: http.MethodGet, Path: "/api/v1/admin/articles", Summary: "list articles"}, func(ctx context.Context, input *articleListRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		result, err := service.List(ctx, actor, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("article list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-get", Method: http.MethodGet, Path: "/api/v1/admin/articles/{id}", Summary: "get article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid article input")}, nil
		}
		item, err := service.Get(ctx, actor, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article detail", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-update", Method: http.MethodPatch, Path: "/api/v1/admin/articles/{id}", Summary: "update article"}, func(ctx context.Context, input *articleUpdateRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid article input")}, nil
		}
		item, err := service.Update(ctx, actor, UpdateInput{ID: id, Title: stringValue(input.Body.Title), Summary: input.Body.Summary, Content: input.Body.Content, CategoryID: input.Body.CategoryID})
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article updated", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-delete", Method: http.MethodDelete, Path: "/api/v1/admin/articles/{id}", Summary: "delete article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid article input")}, nil
		}
		if err := service.Delete(ctx, actor, id); err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article deleted", map[string]uint{"id": id})}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-publish", Method: http.MethodPost, Path: "/api/v1/admin/articles/{id}/publish", Summary: "publish article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid article input")}, nil
		}
		item, err := service.Publish(ctx, actor, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article published", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-article-unpublish", Method: http.MethodPost, Path: "/api/v1/admin/articles/{id}/unpublish", Summary: "unpublish article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
		actor, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid article input")}, nil
		}
		item, err := service.Unpublish(ctx, actor, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &articleEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &articleEnvelopeOutput{Status: http.StatusOK, Body: response.OK("article unpublished", item)}, nil
	})
}
