package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

type routeService interface {
	Login(ctx context.Context, username, password string) (*LoginResult, error)
}

type loginInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
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

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "login and get jwt",
	}, func(ctx context.Context, input *loginInput) (*envelopeOutput, error) {
		result, err := service.Login(ctx, input.Body.Username, input.Body.Password)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("login success", result)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-me",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/me",
		Summary:     "get current user",
	}, func(ctx context.Context, input *struct{}) (*envelopeOutput, error) {
		currentUser, ok := CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &envelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("current user", currentUser)}, nil
	})
}
