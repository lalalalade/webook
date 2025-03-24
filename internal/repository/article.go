package repository

import (
	"context"
	"github.com/lalalalade/webook/internal/domain"
	dao "github.com/lalalalade/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticlesStatus) error
}

type CacheArticleRepository struct {
	dao dao.ArticleDAO

	// v1 操作两个dao
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果要在 repo 上操作事务
	// 那么就只能利用 db 开启事务后，创建基于事务的 DAO
	// 或者，直接去除 DAO 这一层，在 repo 的实现中，直接操作 db（不推荐）
	db *gorm.DB
}

func NewArticleRepository(dao dao.ArticleDAO) ArticleRepository {
	return &CacheArticleRepository{
		dao: dao,
	}
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CacheArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	if id > 0 {
		err = c.authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = c.authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	// 操作线上库，保存数据，同步过来
	// INSERT or UPDATE
	// 数据库有则更新，没有则插入
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

// SyncV2 尝试在 repo 层解决事务问题
func (c *CacheArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	// 利用 tx 构建dao
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var (
		id  = art.Id
		err error
	)
	artn := c.toEntity(art)
	if id > 0 {
		err = author.UpdateById(ctx, artn)
	} else {
		id, err = author.Insert(ctx, artn)
	}
	if err != nil {
		// 执行有问题 回滚
		//tx.Rollback()
		return 0, err
	}
	// 操作线上库，保存数据，同步过来
	// INSERT or UPDATE
	// 数据库有则更新，没有则插入
	err = reader.UpsertV2(ctx, dao.PublishedArticle(artn))
	// 执行成功 提交
	tx.Commit()
	return id, err
}

func (c *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

func (c *CacheArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticlesStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c *CacheArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
