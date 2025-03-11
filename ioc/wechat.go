package ioc

import "github.com/lalalalade/webook/internal/service/oauth2/wechat"

func InitOAuth2WechatService() wechat.Service {
	appId := "asdfbjqe"
	appSecret := "asdfjkl"
	return wechat.NewService(appId, appSecret)
}
