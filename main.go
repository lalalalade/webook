package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/lalalalade/webook/config"
	"github.com/lalalalade/webook/internal/repository"
	"github.com/lalalalade/webook/internal/repository/dao"
	"github.com/lalalalade/webook/internal/service"
	"github.com/lalalalade/webook/internal/web"
	"github.com/lalalalade/webook/internal/web/middleware"
	"github.com/lalalalade/webook/pkg/ginx/middlewares/ratelimit"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()

	u := initUser(db)

	server := initWebServer()

	u.RegisterRoutes(server)
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//	ctx.String(http.StatusOK, "hello world")
	//})
	server.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
	server.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 让前端能读到
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	//store := cookie.NewStore([]byte("secret"))
	store := memstore.NewStore([]byte("7aB3rR9qFyZx6TgKpL8HjD2N4vM5cW1sV"),
		[]byte("Xk9Lm4nB7vR2qZ8tYw3pD6sF1gH5jKcV"))
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("7aB3rR9qFyZx6TgKpL8HjD2N4vM5cW1sV"), []byte("Xk9Lm4nB7vR2qZ8tYw3pD6sF1gH5jKcV"))
	server.Use(sessions.Sessions("mysession", store))

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").
		Build())
	return server
}
