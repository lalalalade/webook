package service

import (
	"context"
	"github.com/lalalalade/webook/internal/domain"
	"github.com/lalalalade/webook/internal/repository"
	"github.com/lalalalade/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	// ListPub 根据更新时间来分页，更新时间必须小于 startTime
	ListPub(ctx context.Context, startTime time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository

	// V1 依靠两个不同的 repo 来解决这种跨表，跨库的问题
	author repository.ArticleAuthorRepository
	reader repository.ArticleReaderRepository
	l      logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(author repository.ArticleAuthorRepository,
	reader repository.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnPublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (a *articleService) update(ctx context.Context, art domain.Article) error {
	//artInDB := a.repo.FindById(ctx, art.Id)
	//if art.Author.Id != artInDB.Author.Id {
	//	return errors.New("更新别人的数据")
	//}
	return a.repo.Update(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.author.Update(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("部分失败，保存到线上库失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，重试彻底失败",
			logger.Int64("art_id", art.Id),
			logger.Error(err))
		// 接入告警系统，手工处理
		// 走异步，直接保存到本地文件
		// 走 canal
		// 走 mq
	}
	return id, err
}

func (a *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetPublishedById(ctx, id)
}

func (a *articleService) ListPub(ctx context.Context,
	startTime time.Time,
	offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, startTime, offset, limit)
}
