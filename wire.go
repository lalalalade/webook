//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lalalalade/webook/internal/repository"
	"github.com/lalalalade/webook/internal/repository/cache"
	"github.com/lalalalade/webook/internal/repository/dao"
	"github.com/lalalalade/webook/internal/repository/dao/article"
	"github.com/lalalalade/webook/internal/service"
	"github.com/lalalalade/webook/internal/web"
	ijwt "github.com/lalalalade/webook/internal/web/jwt"
	"github.com/lalalalade/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,

		// 初始化 DAO
		dao.NewUserDAO, article.NewGORMArticleDAO,

		cache.NewUserCache, cache.NewCodeCache,

		repository.NewUserRepository, repository.NewCodeRepository, repository.NewArticleRepository,

		service.NewUserService, service.NewCodeService, service.NewArticleService,
		ioc.InitSMSService, ioc.InitOAuth2WechatService, ioc.NewWechatHandlerConfig,

		web.NewUserHandler, web.NewOAuth2WechatHandler, web.NewArticleHandler, ijwt.NewRedisJWTHandler,
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
