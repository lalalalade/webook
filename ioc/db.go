package ioc

import (
	"fmt"
	"github.com/lalalalade/webook/internal/repository/dao"
	"github.com/lalalalade/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

// InitDB 初始化数据库
func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		DSN: "root:root@tcp(localhost:13316)/webook",
	}
	err := viper.UnmarshalKey("db.mysql", &cfg)
	if err != nil {
		panic(fmt.Errorf("mysql初始化配置失败: %v \n", err))
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
