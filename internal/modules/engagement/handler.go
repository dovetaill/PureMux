package engagement

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

type routeService interface {
	Like(ctx context.Context, memberID, articleID uint) (*Like, error)
	Unlike(ctx context.Context, memberID, articleID uint) error
	Favorite(ctx context.Context, memberID, articleID uint) (*Favorite, error)
	Unfavorite(ctx context.Context, memberID, articleID uint) error
	ListFavorites(ctx context.Context, memberID uint, page, pageSize int) (*FavoritesResult, error)
}

type articleIDRequest struct {
	ID string `path:"id"`
}

type favoritesListRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type envelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "engagement-like-create", Method: http.MethodPost, Path: "/api/v1/articles/{id}/likes", Summary: "like article"}, func(ctx context.Context, input *articleIDRequest) (*envelopeOutput, error) {
		memberID, status, body, ok := currentMember(ctx)
		if !ok {
			return &envelopeOutput{Status: status, Body: body}, nil
		}
		articleID, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid engagement input")}, nil
		}
		item, err := service.Like(ctx, memberID, articleID)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusCreated, Body: response.OK("article liked", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "engagement-like-delete", Method: http.MethodDelete, Path: "/api/v1/articles/{id}/likes", Summary: "remove article like"}, func(ctx context.Context, input *articleIDRequest) (*envelopeOutput, error) {
		memberID, status, body, ok := currentMember(ctx)
		if !ok {
			return &envelopeOutput{Status: status, Body: body}, nil
		}
		articleID, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid engagement input")}, nil
		}
		if err := service.Unlike(ctx, memberID, articleID); err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("article like removed", map[string]uint{"member_id": memberID, "article_id": articleID})}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "engagement-favorite-create", Method: http.MethodPost, Path: "/api/v1/articles/{id}/favorites", Summary: "favorite article"}, func(ctx context.Context, input *articleIDRequest) (*envelopeOutput, error) {
		memberID, status, body, ok := currentMember(ctx)
		if !ok {
			return &envelopeOutput{Status: status, Body: body}, nil
		}
		articleID, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid engagement input")}, nil
		}
		item, err := service.Favorite(ctx, memberID, articleID)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusCreated, Body: response.OK("article favorited", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "engagement-favorite-delete", Method: http.MethodDelete, Path: "/api/v1/articles/{id}/favorites", Summary: "remove article favorite"}, func(ctx context.Context, input *articleIDRequest) (*envelopeOutput, error) {
		memberID, status, body, ok := currentMember(ctx)
		if !ok {
			return &envelopeOutput{Status: status, Body: body}, nil
		}
		articleID, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid engagement input")}, nil
		}
		if err := service.Unfavorite(ctx, memberID, articleID); err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("article favorite removed", map[string]uint{"member_id": memberID, "article_id": articleID})}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "engagement-favorite-list", Method: http.MethodGet, Path: "/api/v1/me/favorites", Summary: "list my favorites"}, func(ctx context.Context, input *favoritesListRequest) (*envelopeOutput, error) {
		memberID, status, body, ok := currentMember(ctx)
		if !ok {
			return &envelopeOutput{Status: status, Body: body}, nil
		}
		result, err := service.ListFavorites(ctx, memberID, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.Paged("favorite list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})
}

func currentMember(ctx context.Context) (uint, int, response.Envelope, bool) {
	currentUser, ok := authmodule.CurrentUserFromContext(ctx)
	if !ok {
		status := http.StatusUnauthorized
		return 0, status, response.Fail(status, "unauthorized"), false
	}
	return currentUser.ID, 0, response.Envelope{}, true
}

func parseID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}
