package auth

import (
	"context"
	"time"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
	"github.com/dovetaill/PureMux/pkg/database"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	Role         string    `gorm:"size:16;not null;index" json:"role"`
	Status       string    `gorm:"size:16;not null;index" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type CurrentUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

func (u *User) ToCurrentUser() CurrentUser {
	if u == nil {
		return CurrentUser{}
	}
	return CurrentUser{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
		Status:   u.Status,
	}
}

type currentUserContextKey struct{}

func ContextWithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserContextKey{}, user)
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	user, ok := ctx.Value(currentUserContextKey{}).(CurrentUser)
	return user, ok
}

func init() {
	bootstrap.RegisterBusinessModels(User{})
	bootstrap.RegisterSeedAdminSupport(func(resources *database.Resources) bootstrap.SeedAdminStore {
		if resources == nil || resources.MySQL == nil {
			return nil
		}
		return NewRepository(resources.MySQL)
	}, HashPassword)
}
