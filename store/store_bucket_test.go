package store

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	"kdqueue/initial"
	_ "kdqueue/xtesting"
)

func TestBucketStore_Add(t *testing.T) {

	var (
		topic = `mock-topic`
		body  = `{"a":"b","c":10}`
		ctx   = context.Background()
	)

	s := NewBucketStore(func(opt *BucketStoreOptions) {
		opt.BucketCount = 10
	})

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

		const msgCount = 10

		var ids []uint64
		for i := 1; i <= msgCount; i++ {
			So(s.Add(ctx, topic, uint64(i), Message{
				DelayMs:     100,
				Body:        body,
				CreatedAtMs: time.Now().UnixNano() / 1e6,
			}), ShouldBeNil)
			ids = append(ids, uint64(i))
		}

		var msgs []Message
		So(db.Find(&msgs).Error, ShouldBeNil)
		So(len(msgs), ShouldEqual, msgCount)
		for _, msg := range msgs {
			So(msg.Topic, ShouldEqual, topic)
			So(msg.Body, ShouldEqual, body)
		}

		Convey("test delete", func() {
			for _, id := range ids {
				So(s.Delete(ctx, topic, id), ShouldBeNil)
			}
			var msg Message
			So(db.First(&msg).Error, ShouldBeNil)
		})

	})
}

func TestBucketStore_FetchDelayMessage(t *testing.T) {

	var (
		topic = `mock-topic`
		body  = `{"a":"b","c":10}`
		ctx   = context.Background()
	)

	s := NewBucketStore(func(opt *BucketStoreOptions) {
		opt.BucketCount = 10
	})

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

		const msgCount = 10

		var ids []uint64
		for i := 1; i <= msgCount; i++ {
			So(s.Add(ctx, topic, uint64(i), Message{
				DelayMs:     100,
				Body:        body,
				CreatedAtMs: time.Now().UnixNano() / 1e6,
			}), ShouldBeNil)
			ids = append(ids, uint64(i))
		}

		<-time.After(100 * time.Millisecond)

		var wg sync.WaitGroup
		wg.Add(msgCount)
		go func() {
			s.FetchDelayMessage(ctx, func(t string, i uint64, m Message) error {
				defer wg.Done() // need receive enough message
				return nil
			})
		}()
		wg.Wait()
		<-time.After(time.Millisecond * 100)

		var msg Message
		So(db.Where("status=?", StatusDelay).First(&msg).Error, ShouldEqual, gorm.ErrRecordNotFound)

		msg = Message{}
		So(db.Where("status=?", StatusSpent).First(&msg).Error, ShouldBeNil)
	})
}
