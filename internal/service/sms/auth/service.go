package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lalalalade/webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

// Send 发送短信 biz必须是线下沟通的token
func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {

	var tc Claims
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token不合法")
	}
	return s.svc.Send(ctx, biz, args, numbers...)
}

type Claims struct {
	TplId string
	jwt.RegisteredClaims
}
