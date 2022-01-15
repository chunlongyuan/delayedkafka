package ha

// mock
//go:generate mockgen -source=ha.go -destination=ha_mock.go -package=ha

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"kdqueue/locker"
	"kdqueue/share/xid"
)

const (
	kdQueueHA = "kdqueue/ha"
)

var (
	ErrNotMaster = errors.New("not master")
)

// HA 负责管理主备
type HA interface {
	// MushMaster 通过分布式锁让锁的持有者作为 Master 节点工作
	MushMaster(context.Context) error
}

type ha struct {
	nodeId string
	once   sync.Once
	locker locker.Locker
}

func NewHA(opts ...Option) HA {

	opt := Options{Locker: locker.NewRedisLocker(), NodeId: strconv.Itoa(int(xid.Get() % 1000))}

	for _, o := range opts {
		o(&opt)
	}

	return &ha{locker: opt.Locker, nodeId: opt.NodeId}
}

func (p *ha) MushMaster(ctx context.Context) error {
	// 争抢 master, 争抢到之后通过心跳将其长期持有
	err := p.locker.Lock(ctx, kdQueueHA, p.nodeId, 5)
	if errors.Is(err, locker.ErrLocked) { // 被别人抢了 master
		return ErrNotMaster
	}
	if err != nil {
		return err
	}
	p.once.Do(func() {
		go p.retainLock(ctx)
	})
	return nil
}

func (p *ha) retainLock(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 4): //
		}
		if err := p.locker.Lock(ctx, kdQueueHA, p.nodeId, 5); err != nil {
			logrus.WithError(err).Errorln("lock err")
		}
	}
}
