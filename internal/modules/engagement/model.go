package engagement

import (
	"time"

	"github.com/dovetaill/PureMux/internal/app/bootstrap"
)

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MemberID  uint      `gorm:"not null;uniqueIndex:idx_member_article_like" json:"member_id"`
	ArticleID uint      `gorm:"not null;uniqueIndex:idx_member_article_like" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string {
	return "article_likes"
}

type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MemberID  uint      `gorm:"not null;uniqueIndex:idx_member_article_favorite" json:"member_id"`
	ArticleID uint      `gorm:"not null;uniqueIndex:idx_member_article_favorite" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Favorite) TableName() string {
	return "article_favorites"
}

func init() {
	bootstrap.RegisterBusinessModels(Like{}, Favorite{})
}
