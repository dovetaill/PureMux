package article

import (
	"time"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

type Article struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:255;not null" json:"title"`
	Summary     string     `gorm:"type:text" json:"summary"`
	Content     string     `gorm:"type:longtext;not null" json:"content"`
	Status      string     `gorm:"size:16;not null;index" json:"status"`
	AuthorID    uint       `gorm:"not null;index" json:"author_id"`
	CategoryID  uint       `gorm:"not null;index" json:"category_id"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Article) TableName() string {
	return "articles"
}

type ListFilter struct {
	AuthorID *uint
}

func init() {
	bootstrap.RegisterBusinessModels(Article{})
}
