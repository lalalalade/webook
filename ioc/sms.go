package ioc

import (
	"github.com/lalalalade/webook/internal/service/sms"
	"github.com/lalalalade/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	svc := memory.NewService()

	return svc
}
