package dao

import (
	"github.com/lalalalade/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &article.Article{}, &article.PublishArticle{})
}
