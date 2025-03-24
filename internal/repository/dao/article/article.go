package article

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 只更新 标题、内容、状态
func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	art.Utime = time.Now().UnixMilli()
	// 依赖 gorm 忽略零值的特性，会用主键进行更新
	// 可读性很差
	res := dao.db.WithContext(ctx).Model(&Article{}).
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
	return nil
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) ListPubByUtime(ctx context.Context, utime time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	err := dao.db.WithContext(ctx).Order("utime DESC").
		Where("utime < ?", utime.UnixMilli()).
		Limit(limit).Offset(offset).Find(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDAO := NewGORMArticleDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle(art)
	publishArt.Utime = now
	publishArt.Ctime = now
	err = tx.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
}

func (dao *GORMArticleDAO) SyncClosure(ctx context.Context,
	art Article) (int64, error) {
	var (
		id = art.Id
	)
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := art
		publishArt.Utime = now
		publishArt.Ctime = now
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   art.Title,
				"content": art.Content,
				"utime":   now,
			}),
		}).Create(&publishArt).Error
	})
	return id, err
}

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticle) error {
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
		return tx.Model(&PublishedArticle{}).Where("id=?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
}
