package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/lalalalade/webook/internal/domain"
	"github.com/lalalalade/webook/internal/repository/cache"
	dao "github.com/lalalalade/webook/internal/repository/dao/article"
	"github.com/lalalalade/webook/pkg/logger"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticlesStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error)
}

type CacheArticleRepository struct {
	dao      dao.ArticleDAO
	userRepo UserRepository
	// v1 操作两个dao
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// 耦合了 DAO 操作的东西
	// 正常情况下，如果要在 repo 上操作事务
	// 那么就只能利用 db 开启事务后，创建基于事务的 DAO
	// 或者，直接去除 DAO 这一层，在 repo 的实现中，直接操作 db（不推荐）
	db *gorm.DB

	cache cache.ArticleCache
	l     logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, l logger.LoggerV1) ArticleRepository {
	return &CacheArticleRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
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
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		c.cache.DelFirstPage(ctx, art.Author.Id)
		er := c.cache.SetPub(ctx, art)
		if er != nil {
			// 不需要特别关心
		}
	}
	return id, err
}

func (c *CacheArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticlesStatus) error {
	return c.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (c *CacheArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 集成复杂的缓存方案
	// 只缓存第一页
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})
	// 回写缓存 可以同步也可以异步
	// 高并发-- Del缓存
	// 不高并发 -- Set缓存
	go func() {
		err = c.cache.SetFirstPage(ctx, uid, data)
		c.l.Error("回写缓存失败", logger.Error(err))
		c.preCache(ctx, data)
	}()
	return data, nil
}

func (c *CacheArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	cacheArt, err := c.cache.Get(ctx, id)
	if err == nil {
		return cacheArt, nil
	}
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *CacheArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}

func (c *CacheArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	// 读取线上库数据
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 组装 user
	usr, err := c.userRepo.FindById(ctx, art.Id)
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticlesStatus(art.Status),
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	go func() {
		if err = c.cache.SetPub(ctx, res); err != nil {
			c.l.Error("缓存已发表文章失败",
				logger.Error(err), logger.Int64("aid", res.Id))
		}
	}()
	return res, nil
}

func (c *CacheArticleRepository) ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error) {
	val, err := c.dao.ListPubByUtime(ctx, utime, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](val, func(idx int, src dao.PublishedArticle) domain.Article {
		// 偷懒写法
		return c.toDomain(dao.Article(src))
	}), nil
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

func (c *CacheArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticlesStatus(art.Status),
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}
