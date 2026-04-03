package article

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

type routeService interface {
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

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "article-create", Method: http.MethodPost, Path: "/api/v1/articles", Summary: "create article"}, func(ctx context.Context, input *articleCreateRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-list", Method: http.MethodGet, Path: "/api/v1/articles", Summary: "list articles"}, func(ctx context.Context, input *articleListRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-get", Method: http.MethodGet, Path: "/api/v1/articles/{id}", Summary: "get article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-update", Method: http.MethodPatch, Path: "/api/v1/articles/{id}", Summary: "update article"}, func(ctx context.Context, input *articleUpdateRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-delete", Method: http.MethodDelete, Path: "/api/v1/articles/{id}", Summary: "delete article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-publish", Method: http.MethodPost, Path: "/api/v1/articles/{id}/publish", Summary: "publish article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
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

	huma.Register(api, huma.Operation{OperationID: "article-unpublish", Method: http.MethodPost, Path: "/api/v1/articles/{id}/unpublish", Summary: "unpublish article"}, func(ctx context.Context, input *articleIDRequest) (*articleEnvelopeOutput, error) {
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
