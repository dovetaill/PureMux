package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/PureMux/internal/api/response"
	"github.com/dovetaill/PureMux/internal/app/bootstrap"
)

type readyOutput struct {
	Body response.Envelope
}

// RegisterReady 注册 /readyz。
func RegisterReady(api huma.API, rt *bootstrap.Runtime) {
	huma.Register(api, huma.Operation{
		OperationID: "readyz",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "readiness check",
	}, func(ctx context.Context, input *struct{}) (*readyOutput, error) {
		deps := map[string]string{
			"database": "down",
			"redis":    "down",
		}
		if rt != nil && rt.Resources != nil {
			if rt.Resources.MySQL != nil {
				deps["database"] = "up"
			}
			if rt.Resources.Redis != nil {
				deps["redis"] = "up"
			}
		}

		return &readyOutput{
			Body: response.OK("ready", map[string]any{
				"dependencies": deps,
			}),
		}, nil
	})
}
