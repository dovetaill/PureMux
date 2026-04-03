package member

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

type selfRouteService interface {
	GetSelf(ctx context.Context, id uint) (*Profile, error)
}

func RegisterSelfRoutes(api huma.API, service selfRouteService) {
	if api == nil || service == nil {
		return
	}

	huma.Register(api, huma.Operation{
		OperationID: "member-self",
		Method:      http.MethodGet,
		Path:        "/api/v1/me",
		Summary:     "get current member profile",
	}, func(ctx context.Context, input *struct{}) (*envelopeOutput, error) {
		currentUser, ok := authmodule.CurrentUserFromContext(ctx)
		if !ok {
			status := http.StatusUnauthorized
			return &envelopeOutput{Status: status, Body: response.Fail(status, "unauthorized")}, nil
		}
		profile, err := service.GetSelf(ctx, currentUser.ID)
		if err != nil {
			status, message := StatusFromError(err)
			return &envelopeOutput{Status: status, Body: response.Fail(status, message)}, nil
		}
		return &envelopeOutput{Status: http.StatusOK, Body: response.OK("member profile", profile)}, nil
	})
}
