package failover

import (
	"context"
	"github.com/lalalalade/webook/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	svcs []sms.Service
	idx  int32
	// 连续超时次数
	cnt int32
	// 阈值
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, cnt, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		cnt:       cnt,
		threshold: threshold,
	}
}
func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddInt32(&t.idx, 1)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		// 切换到下一个服务
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// idx = newIdx
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		// 超时
		atomic.AddInt32(&t.cnt, 1)
	case nil:
		// 没有任何错误，重置计数器
		atomic.StoreInt32(&t.cnt, 0)
	}
	return err
}
