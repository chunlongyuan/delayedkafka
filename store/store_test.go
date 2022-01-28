package store

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	"kdqueue/initial"
	"kdqueue/share/xid"
	_ "kdqueue/xtesting"
)

func Test_store_AddAndDelete(t *testing.T) {

	var (
		id          = xid.Get()
		topic       = `mock-topic`
		body        = `{"a":"b","c":10}`
		ctx         = context.Background()
		delayMs     = int64(1000)
		createdAtMs = time.Now().UnixNano() / 1e6
	)

	s := NewStore().(*store)

	Convey("test add", t, func() {

		db := initial.DefDB
		rc := initial.DefRedisPool

		conn := rc.Get()
		defer conn.Close()

		So(db.AutoMigrate(&Message{}), ShouldBeNil)
		defer func() {
			db.Exec(`drop table ` + TableMessage)
			conn.Do("FLUSHALL")
		}()

		var msg Message
		So(db.First(&msg).Error, ShouldEqual, gorm.ErrRecordNotFound)

		hLen, err := redigo.Int(conn.Do(`HLEN`, s.getHashStoreKey()))
		So(err, ShouldBeNil)
		So(hLen, ShouldBeZeroValue)

		zLen, err := redigo.Int(conn.Do(`ZCOUNT`, s.getZSetStoreKey(), "-inf", "+inf"))
		So(err, ShouldBeNil)
		So(zLen, ShouldBeZeroValue)

		message := Message{
			ID:          id,
			Topic:       topic,
			DelayMs:     delayMs,
			Body:        body,
			CreatedAtMs: createdAtMs,
		}

		So(s.Add(ctx, topic, id, message), ShouldBeNil)

		msg = Message{}
		So(db.First(&msg).Error, ShouldBeNil)
		So(msg.Body, ShouldEqual, body)
		So(msg.DelayMs, ShouldEqual, delayMs)
		So(msg.CreatedAtMs, ShouldEqual, createdAtMs)
		So(msg.ID, ShouldEqual, id)
		So(msg.Topic, ShouldEqual, topic)
		So(msg.TTRms, ShouldEqual, time.Unix(0, createdAtMs*1e6).Add(time.Duration(delayMs)*1e6).UnixNano()/1e6)

		hLen, err = redigo.Int(conn.Do(`HLEN`, s.getHashStoreKey()))
		So(err, ShouldBeNil)
		So(hLen, ShouldEqual, 1)

		now := time.Now().Add(time.Second).UnixNano() / 1e6
		ids, err := redigo.Strings(conn.Do(`ZRANGEBYSCORE`, s.getZSetStoreKey(), `-inf`, now, `LIMIT`, 0, 10))
		So(err, ShouldBeNil)
		So(ids, ShouldResemble, []string{strconv.FormatUint(id, 10)})

		zLen, err = redigo.Int(conn.Do(`ZCOUNT`, s.getZSetStoreKey(), "-inf", "+inf"))
		So(err, ShouldBeNil)
		So(zLen, ShouldEqual, 1)

		jsonBytes, err := json.Marshal(&message)
		So(err, ShouldBeNil)
		bodies, err := redigo.Strings(conn.Do(`HMGET`, redigo.Args{}.Add(s.getHashStoreKey()).AddFlat(ids)...))
		So(err, ShouldBeNil)
		So(bodies, ShouldResemble, []string{string(jsonBytes)})

		Convey("test delete", func() {

			So(s.Delete(ctx, topic, id), ShouldBeNil)

			hLen, err = redigo.Int(conn.Do(`HLEN`, s.getHashStoreKey()))
			So(err, ShouldBeNil)
			So(hLen, ShouldEqual, 0)

			zLen, err = redigo.Int(conn.Do(`ZCOUNT`, s.getZSetStoreKey(), "-inf", "+inf"))
			So(err, ShouldBeNil)
			So(zLen, ShouldEqual, 0)

			msg = Message{}
			So(db.Where(`id=? and state=1`, id).First(&msg).Error, ShouldBeNil)

			msg = Message{}
			So(db.Where(`id=? and state=0`, id).First(&msg).Error, ShouldEqual, gorm.ErrRecordNotFound)
		})
	})
}

func Test_store_FetchDelayMessage(t *testing.T) {

	var (
		id      = xid.Get()
		topic   = `mock-topic`
		body    = `{"a":"b","c":10}`
		ctx     = context.Background()
		delayMs = int64(100)
	)

	s := NewStore().(*store)

	Convey("test fetch success", t, func() {

		db := initial.DefDB
		rc := initial.DefRedisPool

		conn := rc.Get()
		defer conn.Close()

		So(db.AutoMigrate(&Message{}), ShouldBeNil)
		defer func() {
			db.Exec(`drop table ` + TableMessage)
			conn.Do("FLUSHALL")
		}()

		message := Message{
			DelayMs:     delayMs,
			Body:        body,
			CreatedAtMs: time.Now().UnixNano() / 1e6,
		}
		So(s.Add(ctx, topic, id, message), ShouldBeNil)

		<-time.After(time.Duration(delayMs) * time.Millisecond)

		var wg sync.WaitGroup
		wg.Add(1)
		err := s.FetchDelayMessage(ctx, func(t string, i uint64, m Message) error {
			defer wg.Done()

			So(t, ShouldEqual, topic)
			So(i, ShouldEqual, id)
			So(body, ShouldEqual, m.Body)

			return nil
		})
		So(err, ShouldResemble, ErrNoData)
		wg.Wait()

		var msg Message
		So(db.Where("state=?", StatusDelay).First(&msg).Error, ShouldEqual, gorm.ErrRecordNotFound)

		msg = Message{}
		So(db.Where("state=?", StatusSpent).First(&msg).Error, ShouldBeNil)

	})
}
