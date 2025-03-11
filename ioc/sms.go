package ioc

import (
	"github.com/lalalalade/webook/internal/service/sms"
	"github.com/lalalalade/webook/internal/service/sms/memory"
	"github.com/redis/go-redis/v9"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	//svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
	//	ratelimit2.NewRedisSlideWindowLimiter(cmd, time.Second, 100))
	//
	//return retryable.NewService(svc, 3)
	return memory.NewService()
}
