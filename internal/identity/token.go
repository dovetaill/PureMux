package identity

import (
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
	"github.com/dovetaill/PureMux/pkg/config"
)

type TokenManager = authmodule.TokenManager
type TokenClaims = authmodule.Claims

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return authmodule.NewTokenManager(cfg)
}
