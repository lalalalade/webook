package article

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, dao Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	art.Utime = time.Now().UnixMilli()
	// 依赖 gorm 忽略零值的特性，会用主键进行更新
	// 可读性很差
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id=? AND author_id", art.Id).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}
	return res.Error
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id = art.Id
	)
	// 先操作制作表 再操作线上表
	// 在事务内部 采用闭包形态
	// GORM 帮我们管理了事务的生命周期
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 操作线上库
		return txDAO.Upsert(ctx, PublishArticle{Article: art})
	})
	return id, err
}

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Clauses(clause.OnConflict{
		// 哪些列冲突
		//Columns: []clause.Column{{Name: "id"}},
		// 意思是数据冲突，啥也不干
		//DoNothing:
		// 数据冲突了，并且符合where条件的 就会执行 DoUpdates
		//Where:
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&art).Error
	return err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id=? AND author_id=?", id, author).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么id错，要么作者不对
			// 后者情况下，小心攻击
			return errors.New("误操作非自己的文章")
		}
		return tx.Model(&PublishArticle{}).Where("id=?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
}

// Article 制作库表
type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index"`

	Status uint8
	Ctime  int64
	Utime  int64
}
