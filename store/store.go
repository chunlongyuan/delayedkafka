package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"strconv"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"kdqueue/config"
	"kdqueue/initial"
	"kdqueue/share/sqlerr"
)

// 负责存储

type Store interface {
	Add(ctx context.Context, topic string, id uint64, msg Message) error
	Delete(ctx context.Context, topic string, id uint64) error
	FetchDelayMessage(ctx context.Context, handle func(topic string, id uint64, msg Message) error) error
}

var ErrNoData = errors.New(`message not found`)

type store struct {
	key      string //
	db       *gorm.DB
	rPool    *redigo.Pool
	perCount int
}

func NewStore(opts ...Option) Store {

	opt := Options{Key: `default`, DB: initial.DefDB, Pool: initial.DefRedisPool, PerCount: 10}

	for _, o := range opts {
		o(&opt)
	}

	if opt.PerCount < 1 {
		panic(fmt.Sprintf(`illegal perCount: %v`, opt.PerCount))
	}

	return &store{key: opt.Key, db: opt.DB, rPool: opt.Pool, perCount: opt.PerCount}
}

func (p *store) Add(ctx context.Context, topic string, id uint64, msg Message) error {

	logger := logrus.WithField("topic", topic).WithField("id", id).WithField("msg", msg)

	logger.Infoln("add message")

	// 开启 db 事务
	tx := p.db.Begin()

	defer func() {
		if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.WithError(err).Errorln("rollback err")
		}
	}()

	conn := p.rPool.Get()
	defer conn.Close()

	// save data to mysql message table
	infraMsg := Message{
		ID:          id,
		Topic:       topic,
		TTRms:       time.Unix(0, msg.CreatedAtMs*1e6).Add(time.Duration(msg.DelayMs)*time.Millisecond).UnixNano() / 1e6,
		DelayMs:     msg.DelayMs,
		Body:        msg.Body,
		CreatedAtMs: msg.CreatedAtMs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	infraMsgBytes, err := json.Marshal(&infraMsg)
	if err != nil {
		logger.WithError(err).Errorln(`marshal err`)
		return err
	}

	if err := tx.Table(TableMessage).Create(&infraMsg).Error; err != nil && !sqlerr.IsDuplicateKey(err) {
		logger.WithError(err).Errorln("create message err")
		return err
	}

	idStr := strconv.FormatUint(id, 10)

	// save data to redis hmap and save id to redis zset
	script := `
redis.call('HSETNX', KEYS[1], ARGV[1], ARGV[2])
redis.call('ZADD', KEYS[2], ARGV[3], ARGV[4])`
	if _, err = conn.Do("EVAL", script, 2, p.getHashStoreKey(), p.getZSetStoreKey(), idStr, string(infraMsgBytes), infraMsg.TTRms, idStr); err != nil {
		logger.WithError(err).Errorf("redis err by script: %v", script)
		return err
	}
	logger.Infof("redis ok by script: %v", script)

	return tx.Commit().Error
}

func (p *store) Delete(ctx context.Context, topic string, id uint64) error {

	logger := logrus.WithField("topic", topic).WithField("id", id)

	logger.Infoln("delete message")

	// 开启 db 事务
	tx := p.db.Begin()

	defer func() {
		if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.WithError(err).Errorln("rollback err")
		}
	}()

	conn := p.rPool.Get()
	defer conn.Close()

	// 删除 db
	if err := tx.Exec("update "+TableMessage+" set state=? where id=?", StatusDelete, id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.WithError(err).Errorln("delete db err")
		return err
	}

	idStr := strconv.FormatUint(id, 10)
	// 删除 hmap 和 zset
	script := `
redis.call('HDEL', KEYS[1], ARGV[1])
redis.call('ZREM', KEYS[2], ARGV[2])
`
	if _, err := conn.Do("EVAL", script, 2, p.getHashStoreKey(), p.getZSetStoreKey(), idStr, idStr); err != nil {
		logger.WithError(err).Errorf("redis err by script: %v", script)
		return err
	}
	logger.Infof("redis ok by script: %v", script)

	return tx.Commit().Error
}

func (p *store) FetchDelayMessage(ctx context.Context, handle func(topic string, id uint64, msg Message) error) error {

	logger := logrus.WithContext(ctx)

	conn := p.rPool.Get()
	defer conn.Close()

	// find zset
	now := time.Now().UnixNano() / 1e6
	ids, err := redigo.Strings(conn.Do(`ZRANGEBYSCORE`, p.getZSetStoreKey(), `-inf`, now, `LIMIT`, 0, p.perCount))
	if err != nil {
		logger.WithError(err).Errorln("range by score err")
		return err
	}
	if len(ids) == 0 { // 没有到期的
		logger.Infoln("range by score got empty")
		return ErrNoData
	}

	// find hmap
	infraMessages, err := redigo.Strings(conn.Do(`HMGET`, redigo.Args{p.getHashStoreKey()}.AddFlat(ids)...))
	if err != nil {
		logger.WithError(err).Errorln("hmget err")
		return err
	}

	if len(infraMessages) == 0 {
		// 有 id 但没有内容的数据移除掉
		if _, err = conn.Do(`ZREM`, redigo.Args{p.getZSetStoreKey()}.AddFlat(ids)...); err != nil {
			logger.WithError(err).Errorln("zrem err")
		}
		return ErrNoData
	}

	handleBodyFun := func(id, infraMessage string) error {

		// 开启 db 事务
		tx := p.db.Begin()

		defer func() {
			if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
				logger.WithError(err).Errorln("rollback err")
			}
		}()

		var msg Message
		if err := json.Unmarshal([]byte(infraMessage), &msg); err != nil {
			logger.WithError(err).Errorln("unmarshal err")
		} else { // 解析成功才抛上去
			if err := handle(msg.Topic, msg.ID, Message{DelayMs: msg.DelayMs, Body: msg.Body, CreatedAtMs: msg.CreatedAtMs}); err != nil {
				logger.WithError(err).WithField("msg", msg).Errorln("handle msg err")
				return err // 只有 handle 错误才认为出错, 并且该错误不能将消息消费掉
			}
			logger.WithField("message", msg).Debugln("spent msg")
		}

		// remove from hmap and zset
		script := `
redis.call('HDEL', KEYS[1], ARGV[1])
redis.call('ZREM', KEYS[2], ARGV[2])
`
		_, err := conn.Do("EVAL", script, 2, p.getHashStoreKey(), p.getZSetStoreKey(), id, id)
		if err != nil {
			logger.WithError(err).Errorf("redis err by script: %v", script)
			return err
		}
		logger.Infof("redis ok by script: %v", script)

		// update db state
		if err := tx.Exec(`update `+TableMessage+` set state=? where id=?`, StatusSpent, msg.ID).Error; err != nil {
			logger.WithError(err).Errorln(" update db err")
		}

		return tx.Commit().Error
	}

	var validCount int

	for i, infraMessage := range infraMessages {

		if len(infraMessage) == 0 {
			continue
		}

		validCount++

		// 逐条处理
		if err := handleBodyFun(ids[i], infraMessage); err != nil {
			return err
		}
	}

	if validCount < p.perCount { //  有效数据不足一页
		return ErrNoData
	}

	return nil
}

func (p *store) getHashStoreKey() string {
	return fmt.Sprintf(`%s/hash/%s`, config.Cfg.QueueKeyword, p.key)
}

func (p *store) getZSetStoreKey() string {
	return fmt.Sprintf(`%s/zset/%s`, config.Cfg.QueueKeyword, p.key)
}
