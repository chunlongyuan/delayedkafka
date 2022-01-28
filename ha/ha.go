package ha

// mock
//go:generate mockgen -source=ha.go -destination=ha_mock.go -package=ha

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"dk/config"
	"dk/locker"
	"dk/share/ip"
	"dk/share/xid"
)

var (
	ErrNotMaster = errors.New("not master")
	//
	dkHA string
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

	dkHA = fmt.Sprintf("%s/ha", config.Cfg.QueueKeyword)

	nodeId := genNodeId()
	fmt.Printf("ha: nodeId:%v\n", nodeId)

	opt := Options{Locker: locker.NewRedisLocker(), NodeId: nodeId}

	for _, o := range opts {
		o(&opt)
	}

	return &ha{locker: opt.Locker, nodeId: opt.NodeId}
}

func genNodeId() string {
	if len(config.Cfg.NodeId) > 0 {
		return config.Cfg.NodeId
	}
	privateIp := ip.PrivateIPv4()
	if len(privateIp) > 0 {
		return privateIp
	}
	return strconv.Itoa(int(xid.Get() % 1000))
}

func (p *ha) MushMaster(ctx context.Context) error {
	// 争抢 master, 争抢到之后通过心跳将其长期持有
	err := p.locker.Lock(ctx, dkHA, p.nodeId, 5)
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
		if err := p.locker.Lock(ctx, dkHA, p.nodeId, 5); err != nil {
			logrus.WithError(err).Errorln("lock err")
		}
	}
}
