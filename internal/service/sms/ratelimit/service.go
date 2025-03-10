package ratelimit

import (
	"context"
	"fmt"
	"github.com/lalalalade/webook/internal/service/sms"
	"github.com/lalalalade/webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RatelimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 加点新特性 -- 限流
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略
		// 可以不限流：可用性要求高，容错策略
		return fmt.Errorf("短信服务判断是否限流出现问题, %w", err)
	}
	if limited {
		return errLimited
	}

	err = s.svc.Send(ctx, tplId, args, numbers...)
	return err
}
