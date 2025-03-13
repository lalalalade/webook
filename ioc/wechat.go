package ioc

import (
	"github.com/lalalalade/webook/internal/service/oauth2/wechat"
	"github.com/lalalalade/webook/internal/web"
	logger2 "github.com/lalalalade/webook/pkg/logger"
)

// InitOAuth2WechatService 初始化微信OAuth2服务
func InitOAuth2WechatService(l logger2.LoggerV1) wechat.Service {
	appId := "asdfbjqe"
	appSecret := "asdfjkl"
	return wechat.NewService(appId, appSecret, l)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
