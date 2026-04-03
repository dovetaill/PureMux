package member

import (
	"time"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/internal/identity"
	authmodule "github.com/dovetaill/PureMux/internal/modules/auth"
)

const (
	RoleMember = "member"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type Member struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	Status       string    `gorm:"size:16;not null;index" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Member) TableName() string {
	return "members"
}

type Profile struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

func (m *Member) ToCurrentUser() authmodule.CurrentUser {
	if m == nil {
		return authmodule.CurrentUser{}
	}
	return authmodule.CurrentUser{
		ID:       m.ID,
		Username: m.Username,
		Role:     RoleMember,
		Status:   m.Status,
	}
}

func (m *Member) ToActor() identity.Actor {
	currentUser := m.ToCurrentUser()
	return currentUser.ToActor()
}

func (m *Member) ToProfile() Profile {
	currentUser := m.ToCurrentUser()
	return Profile{
		ID:       currentUser.ID,
		Username: currentUser.Username,
		Role:     currentUser.Role,
		Status:   currentUser.Status,
	}
}

func init() {
	bootstrap.RegisterBusinessModels(Member{})
}
