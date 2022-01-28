package store

import (
	"context"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"dk/config"
	"dk/initial"
	"dk/share/xid"
	"dk/xtesting"
	_ "dk/xtesting"
)

func TestSyncer_monitor(t *testing.T) {

	s := NewDatadog(func(opt *SyncerOptions) { opt.MonitordkSeconds = 3 }).(*datadog)

	conn := initial.DefRedisPool.Get()
	defer conn.Close()

	tests := []struct {
		name   string
		before func()
		after  func()
		want   bool
	}{
		{before: func() {}, after: func() {}, want: true},                                            // 首次需要同步
		{before: func() { s.resetMonitor(conn) }, after: func() {}, want: false},                     // 非首次不需要同步
		{before: func() { s.resetMonitor(conn); conn.Do("FLUSHALL") }, after: func() {}, want: true}, // 非首次但数据没了就需要同步
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			defer tt.after()
			defer func() {
				conn.Do("FLUSHALL")
			}()
			if got := s.isNeedSync(); got != tt.want {
				t.Errorf("isNeedSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncer_doSync(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		ctx = context.Background()

		db = initial.DefDB
		rc = initial.DefRedisPool

		topic      = "topic"
		delayMs    = int64(xtesting.RandomInt())
		ttrMs      = int64(xtesting.RandomInt())
		body       = xtesting.RandomStr()
		createAtMs = int64(xtesting.RandomInt())
	)

	Convey("doSync", t, func() {

		conn := rc.Get()
		defer conn.Close()

		So(db.AutoMigrate(&Message{}), ShouldBeNil)
		defer func() {
			db.Exec(`drop table ` + TableMessage)
			conn.Do("FLUSHALL")
		}()

		s := NewDatadog(func(opt *SyncerOptions) { opt.MonitordkSeconds = 3; opt.Store = NewStore() }).(*datadog)
		Convey("sync should success", func() {

			var messages []Message
			for i := 0; i < 3; i++ {
				message := Message{
					ID:          xid.Get(),
					Topic:       topic,
					DelayMs:     delayMs,
					TTRms:       ttrMs,
					State:       int8(i), // state=0 的会被同步到 redis
					Body:        body,
					CreatedAtMs: createAtMs,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				messages = append(messages, message)
			}
			So(db.Create(&messages).Error, ShouldBeNil)

			So(s.doSync(ctx), ShouldBeNil)

			n, err := redis.Int(conn.Do("ZCARD", config.Cfg.QueueKeyword+"/zset/default"))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 1)

			n, err = redis.Int(conn.Do("HLEN", config.Cfg.QueueKeyword+"/hash/default"))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 1)

		})

	})
}
