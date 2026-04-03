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
	Update(ctx context.Context, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id uint) error
}

type createInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Status   string `json:"status"`
	}
}

type listInput struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type updateInput struct {
	ID   string `path:"id"`
	Body struct {
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
		Role     *string `json:"role,omitempty"`
		Status   *string `json:"status,omitempty"`
	}
}

type deleteInput struct {
	ID string `path:"id"`
}

type envelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterRoutes(api huma.API, service routeService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{OperationID: "admin-user-create", Method: http.MethodPost, Path: "/api/v1/admin/users", Summary: "create user"}, func(ctx context.Context, input *createInput) (*envelopeOutput, error) {
		item, err := service.Create(ctx, CreateInput{Username: input.Body.Username, Password: input.Body.Password, Role: input.Body.Role, Status: input.Body.Status})
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusCreated, Body: response.OK("user created", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-user-list", Method: http.MethodGet, Path: "/api/v1/admin/users", Summary: "list users"}, func(ctx context.Context, input *listInput) (*envelopeOutput, error) {
		result, err := service.List(ctx, input.Page, input.PageSize)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("user list", result)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-user-update", Method: http.MethodPut, Path: "/api/v1/admin/users/{id}", Summary: "update user"}, func(ctx context.Context, input *updateInput) (*envelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid user input")}, nil
		}
		item, err := service.Update(ctx, UpdateInput{ID: id, Username: stringValue(input.Body.Username), Password: stringValue(input.Body.Password), Role: stringValue(input.Body.Role), Status: stringValue(input.Body.Status)})
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("user updated", item)}, nil
	})

	huma.Register(api, huma.Operation{OperationID: "admin-user-delete", Method: http.MethodDelete, Path: "/api/v1/admin/users/{id}", Summary: "delete user"}, func(ctx context.Context, input *deleteInput) (*envelopeOutput, error) {
		id, err := parseID(input.ID)
		if err != nil {
			status := http.StatusBadRequest
			return &envelopeOutput{Status: status, Body: response.Fail(status, "invalid user input")}, nil
		}
		if err := service.Delete(ctx, id); err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("user deleted", map[string]uint{"id": id})}, nil
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
