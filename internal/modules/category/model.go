package category

import (
	"time"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
)

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Slug        string    `gorm:"size:128;not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

func init() {
	bootstrap.RegisterBusinessModels(Category{})
}
