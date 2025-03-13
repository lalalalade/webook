package ioc

import (
	"github.com/lalalalade/webook/internal/service/sms"
	"github.com/lalalalade/webook/internal/service/sms/memory"
	"github.com/lalalalade/webook/internal/service/sms/ratelimit"
	"github.com/lalalalade/webook/internal/service/sms/retryable"
	limiter "github.com/lalalalade/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

// InitSMSService 初始化短信服务
func InitSMSService(cmd redis.Cmdable) sms.Service {
	svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
		limiter.NewRedisSlideWindowLimiter(cmd, time.Second, 100))

	return retryable.NewService(svc, 3)
	//return memory.NewService()
}
