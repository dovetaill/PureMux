package identity

import authmodule "github.com/dovetaill/PureMux/internal/modules/auth"

func HashPassword(password string) (string, error) {
	return authmodule.HashPassword(password)
}

func VerifyPassword(hash, password string) error {
	return authmodule.VerifyPassword(hash, password)
}
