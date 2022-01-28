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

	"dk/config"
	"dk/initial"
)

const (
	monitordkValue = 1
)

var (
	halfMonitorTime time.Time
	// SyncState 同步状态
	syncMu    sync.Mutex
	SyncState int
	//
	monitorDKKey string
)

// 负责同步 redis 和 mysql 的数据一致性
// 当redis丢失数据时执行该命令同步数据

type Datadog interface {
	Sync(context.Context) error
}

type datadog struct {
	store            Store
	db               *gorm.DB
	rc               *redis.Pool
	monitorDKSeconds int
	monitorInterval  time.Duration
}

func NewDatadog(opts ...SyncerOption) Datadog {

	monitorDKKey = fmt.Sprintf("%s/monitor/sync", config.Cfg.QueueKeyword)

	opt := SyncerOptions{
		Store:            DefStore,
		DB:               initial.DefDB,
		Redis:            initial.DefRedisPool,
		MonitordkSeconds: 3600,            // one hour
		MonitorInterval:  time.Second * 3, // 检查间隔
	}
	for _, o := range opts {
		o(&opt)
	}
	return &datadog{store: opt.Store, db: opt.DB, rc: opt.Redis, monitorDKSeconds: opt.MonitordkSeconds, monitorInterval: opt.MonitorInterval}
}

func (p *datadog) Sync(ctx context.Context) error {

	logger := logrus.WithContext(ctx)

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
		if p.isNeedSync() {
			logger.Warnln("some problems occurred so need sync data")
			p.doSync(ctx) // 执行同步
		}
		<-time.After(p.monitorInterval)
	}
}

func (p *datadog) isNeedSync() bool {

	logger := logrus.WithField("function", "isNeedSync")

	conn := p.rc.Get()
	defer conn.Close()

	n, err := redis.Int(conn.Do("GET", monitorDKKey))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		logger.WithError(err).Errorln("get err")
		return false
	}

	if n == 0 { // need sync and reset
		logger.Errorln("need sync")
		p.resetMonitor(conn)
		return true
	}

	if time.Now().After(halfMonitorTime) { // 时间过半
		logger.Infoln("over half the time")
		p.resetMonitor(conn)
	}
	return false
}

func (p *datadog) resetMonitor(conn redis.Conn) {

	logger := logrus.WithField("function", "resetMonitor")
	logger.Debugln("reset")

	_, err := conn.Do("SETEX", monitorDKKey, p.monitorDKSeconds, monitordkValue)
	if err != nil {
		logger.WithError(err).Errorln("setex err")
		return
	}
	halfMonitorTime = time.Now().Add(time.Second * time.Duration(p.monitorDKSeconds/2))
}

func (p *datadog) doSync(ctx context.Context) error {

	var total int
	defer func() {
		logrus.Warnf("do sync data count: %v", total)
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
where id>%v and state=%v
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
