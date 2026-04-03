package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

type routeService interface {
	Create(ctx context.Context, input CreateInput) (*User, error)
	List(ctx context.Context, page, pageSize int) (*ListResult, error)
	Get(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id uint) error
}

type userCreateBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type userCreateRequest struct {
	Body userCreateBody
}

type userListRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type userIDRequest struct {
	ID string `path:"id"`
}

type userUpdateBody struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Role     *string `json:"role,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type userUpdateRequest struct {
	ID   string `path:"id"`
	Body userUpdateBody
}

type userEnvelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "admin-user-create", Method: http.MethodPost, Path: "/api/v1/admin/users", Summary: "create user"}, func(ctx context.Context, input *userCreateRequest) (*userEnvelopeOutput, error) {
		item, err := service.Create(ctx, CreateInput{Username: input.Body.Username, Password: input.Body.Password, Role: input.Body.Role, Status: input.Body.Status})
		if err != nil {
			status, message := StatusFromError(err)
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &userEnvelopeOutput{Status: http.StatusCreated, Body: response.OK("user created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-user-list", Method: http.MethodGet, Path: "/api/v1/admin/users", Summary: "list users"}, func(ctx context.Context, input *userListRequest) (*userEnvelopeOutput, error) {
		result, err := service.List(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &userEnvelopeOutput{Status: http.StatusOK, Body: response.Paged("user list", result.Page, result.PageSize, result.Total, result.Items)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-user-get", Method: http.MethodGet, Path: "/api/v1/admin/users/{id}", Summary: "get user"}, func(ctx context.Context, input *userIDRequest) (*userEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid user input")}, nil
		}
		item, err := service.Get(ctx, id)
		if err != nil {
			status, message := StatusFromError(err)
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &userEnvelopeOutput{Status: http.StatusOK, Body: response.OK("user detail", item)}, nil
	})

	userUpdateHandler := func(ctx context.Context, input *userUpdateRequest) (*userEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid user input")}, nil
		}
		item, err := service.Update(ctx, UpdateInput{ID: id, Username: stringValue(input.Body.Username), Password: stringValue(input.Body.Password), Role: stringValue(input.Body.Role), Status: stringValue(input.Body.Status)})
		if err != nil {
			status, message := StatusFromError(err)
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &userEnvelopeOutput{Status: http.StatusOK, Body: response.OK("user updated", item)}, nil
	}

	huma.Register(api, huma.Operation{OperationID: "admin-user-update", Method: http.MethodPut, Path: "/api/v1/admin/users/{id}", Summary: "update user"}, userUpdateHandler)
	huma.Register(api, huma.Operation{OperationID: "admin-user-update-patch", Method: http.MethodPatch, Path: "/api/v1/admin/users/{id}", Summary: "patch user"}, userUpdateHandler)

	huma.Register(api, huma.Operation{OperationID: "admin-user-delete", Method: http.MethodDelete, Path: "/api/v1/admin/users/{id}", Summary: "delete user"}, func(ctx context.Context, input *userIDRequest) (*userEnvelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, "invalid user input")}, nil
		}
		if err := service.Delete(ctx, id); err != nil {
			status, message := StatusFromError(err)
			return &userEnvelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &userEnvelopeOutput{Status: http.StatusOK, Body: response.OK("user deleted", map[string]uint{"id": id})}, nil
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
