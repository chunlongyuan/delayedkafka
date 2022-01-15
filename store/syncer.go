package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"kdqueue/ha"
	"kdqueue/initial"
)

const (
	monitorKDQueueKey   = "kdqueue/monitor/sync"
	monitorKDQueueValue = 1
)

var (
	halfMonitorTime time.Time
	// SyncState 同步状态
	syncMu    sync.Mutex
	SyncState int
)

// 负责同步 redis 和 mysql 的数据一致性
// 当redis丢失数据时执行该命令同步数据

type Syncer interface {
	Sync(context.Context) error
}

type syncer struct {
	ha                    ha.HA
	store                 Store
	db                    *gorm.DB
	rc                    *redis.Pool
	monitorKDQueueSeconds int
	monitorInterval       time.Duration
}

func NewSyncer(opts ...SyncerOption) Syncer {
	opt := SyncerOptions{
		HA:                    ha.NewHA(),
		Store:                 DefStore,
		DB:                    initial.DefDB,
		Redis:                 initial.DefRedisPool,
		MonitorKDQueueSeconds: 3600,            // one hour
		MonitorInterval:       time.Second * 3, // 检查间隔
	}
	for _, o := range opts {
		o(&opt)
	}
	return &syncer{ha: opt.HA, store: opt.Store, db: opt.DB, rc: opt.Redis, monitorKDQueueSeconds: opt.MonitorKDQueueSeconds, monitorInterval: opt.MonitorInterval}
}

func (p *syncer) Sync(ctx context.Context) error {

	logger := logrus.WithContext(ctx)

	// 主节点才处理同步
	if err := p.ha.MushMaster(ctx); err != nil {
		if errors.Is(err, ha.ErrNotMaster) {
			logger.Infoln("slaver node")
			// do nothing for slaver
			return nil
		}
		logger.WithError(err).Errorln("must master err")
		return err
	}

	logger.Infoln("i am master")

	// 设置标记 TTL=2N 秒
	// 每秒查看:
	// 	如果 N 内获取不要数据则启动同步, 并续期 2N 秒
	// 	如果 N 秒还在则直接续期 2N 秒
	for {
		select {
		case <-ctx.Done():
			logger.Warnln("ctx done")
			return nil
		default:
		}
		if p.monitor() {
			logger.Warnln("some problems occurred so need sync data")
			p.doSync(ctx) // 执行同步
		}
		<-time.After(p.monitorInterval)
	}
	return nil
}

func (p *syncer) monitor() bool {

	logger := logrus.WithField("function", "monitor")

	conn := p.rc.Get()
	defer conn.Close()

	n, err := redis.Int(conn.Do("GET", monitorKDQueueKey))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		logger.WithError(err).Errorln("get err")
		return false
	}

	if n == 0 { // need sync and reset
		logger.Errorln("need sync")
		p.resetMonitor(conn)
		return true
	}

	if halfMonitorTime.After(time.Now()) { // 时间过半
		logger.Infoln("over half the time")
		p.resetMonitor(conn)
	}
	return false
}

func (p *syncer) resetMonitor(conn redis.Conn) {

	logger := logrus.WithField("function", "reset monitor")
	logger.Debugln("reset monitor")

	_, err := conn.Do("SETEX", monitorKDQueueKey, p.monitorKDQueueSeconds, monitorKDQueueValue)
	if err != nil {
		logger.WithError(err).Errorln("setex err")
		return
	}
	halfMonitorTime = time.Now().Add(time.Second * time.Duration(p.monitorKDQueueSeconds/2))
}

func (p *syncer) doSync(ctx context.Context) error {

	var total int
	defer func() {
		logrus.Warnln("do sync data count: %v", total)
	}()

	syncMu.Lock()
	SyncState = 1

	defer func() {
		SyncState = 0
		syncMu.Unlock()
	}()

	var (
		lastId  uint64
		maxLoop = 1000 // max loop
	)
	for i := maxLoop; i > 0; i-- {

		sqlStr := fmt.Sprintf(`
select * from %v
where id>%v and status=%v
order by id asc
limit 100
`, TableMessage, lastId, StatusDelay)
		var messages []Message
		if err := p.db.Raw(sqlStr).Scan(&messages).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			logrus.WithError(err).Errorln("scan messages err")
			return err
		}
		if len(messages) == 0 {
			logrus.Infoln("empty messages need sync")
			return nil
		}

		total += len(messages)

		for _, msg := range messages {
			if err := p.store.Add(ctx, msg.Topic, msg.ID, msg); err != nil {
				logrus.WithError(err).Errorln("store add err")
				return err
			}
			lastId = msg.ID
		}
		<-time.After(time.Millisecond * 10)
	}
	logrus.Errorf("exceed max loop %v", maxLoop)
	return nil
}
