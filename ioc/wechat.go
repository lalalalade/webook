package ioc

import (
	"github.com/lalalalade/webook/internal/service/oauth2/wechat"
	"github.com/lalalalade/webook/internal/web"
)

func InitOAuth2WechatService() wechat.Service {
	appId := "asdfbjqe"
	appSecret := "asdfjkl"
	return wechat.NewService(appId, appSecret)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}
