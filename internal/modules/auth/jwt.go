package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dovetaill/PureMux/pkg/config"
	jwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	ttl := time.Duration(cfg.TTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 120 * time.Minute
	}
	issuer := cfg.Issuer
	if issuer == "" {
		issuer = "PureMux"
	}
	return &TokenManager{
		secret: []byte(cfg.Secret),
		issuer: issuer,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *TokenManager) Sign(user CurrentUser) (string, time.Time, error) {
	if m == nil {
		return "", time.Time{}, fmt.Errorf("token manager is required")
	}

	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	if m == nil {
		return nil, ErrUnauthorized
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrUnauthorized
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrUnauthorized
	}
	return claims, nil
}
