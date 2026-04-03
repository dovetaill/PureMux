package identity

import authmodule "github.com/dovetaill/PureMux/internal/modules/auth"

type PrincipalKind string

const (
	PrincipalAdmin  PrincipalKind = "admin"
	PrincipalMember PrincipalKind = "member"
)

type Principal struct {
	Kind     PrincipalKind `json:"kind"`
	UserID   uint          `json:"user_id"`
	Username string        `json:"username"`
	Role     string        `json:"role"`
	Status   string        `json:"status"`
}

func PrincipalFromCurrentUser(user authmodule.CurrentUser) Principal {
	kind := PrincipalMember
	if user.Role == authmodule.RoleAdmin {
		kind = PrincipalAdmin
	}

	return Principal{
		Kind:     kind,
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Status:   user.Status,
	}
}
