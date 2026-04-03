package user

import authmodule "github.com/dovetaill/PureMux/internal/modules/auth"

const (
	RoleAdmin = authmodule.RoleAdmin
	RoleUser  = authmodule.RoleUser

	StatusActive   = authmodule.StatusActive
	StatusDisabled = authmodule.StatusDisabled
)

type User = authmodule.User
