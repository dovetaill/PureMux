package member

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
)

type publicRouteService interface {
	Register(ctx context.Context, username, password string) (*AuthResult, error)
	Login(ctx context.Context, username, password string) (*AuthResult, error)
}

type authBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authRequest struct {
	Body authBody
}

type envelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

func RegisterPublicRoutes(api huma.API, service publicRouteService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{
		OperationID: "member-register",
		Method:      http.MethodPost,
		Path:        "/api/v1/member/auth/register",
		Summary:     "register member and issue token",
	}, func(ctx context.Context, input *authRequest) (*envelopeOutput, error) {
		result, err := service.Register(ctx, input.Body.Username, input.Body.Password)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusCreated, Body: response.OK("member registered", result)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "member-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/member/auth/login",
		Summary:     "login member and issue token",
	}, func(ctx context.Context, input *authRequest) (*envelopeOutput, error) {
		result, err := service.Login(ctx, input.Body.Username, input.Body.Password)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("member login success", result)}, nil
	})
}
