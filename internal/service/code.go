package service

import (
	"context"
	"fmt"
	"github.com/lalalalade/webook/internal/repository"
	"github.com/lalalalade/webook/internal/service/sms"
	"math/rand"
)

const codeTplId = "1865669"

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 生成验证码并发送
func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// 生成验证码
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(100000)
	return fmt.Sprintf("%06d", num)
}
