package repository

import (
	"context"
	"github.com/lalalalade/webook/internal/domain"
)

type ArticleReaderRepository interface {
	// Save 有就更新，没有就新建，upsert的语义
	Save(ctx context.Context, art domain.Article) (int64, error)
}
