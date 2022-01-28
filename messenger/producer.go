package messenger

// mock
//go:generate mockgen -source=producer.go -destination=producer_mock.go -package=delivery

import (
	"context"
	"errors"
	"time"

	"dk/ha"
	"dk/share/xid"
	"dk/store"
)

var (
	DefProducer Producer
)

// Producer 负责生产
type Producer interface {
	// Publish 发布消息, 返回消息 ID
	Publish(ctx context.Context, topic string, msg Message) (uint64, error)
	// Cancel 取消
	Cancel(ctx context.Context, topic string, id uint64) error
}

type producer struct {
	ha       ha.HA
	store    store.Store
	delivery Delivery
}

func NewProducer(opts ...ProducerOption) Producer {

	opt := ProducerOptions{Store: store.DefStore, Delivery: DefDeliver}

	for _, o := range opts {
		o(&opt)
	}
	return &producer{delivery: opt.Delivery, store: opt.Store}
}

func (p *producer) Publish(ctx context.Context, topic string, msg Message) (uint64, error) {

	id := xid.Get()

	now := time.Now()
	message := store.Message{
		ID:          id,
		Topic:       topic,
		TTRms:       time.Unix(0, msg.CreatedAtMs*1e6).Add(time.Duration(msg.DelayMs)*time.Millisecond).UnixNano() / 1e6,
		DelayMs:     msg.DelayMs,
		Body:        msg.Body,
		CreatedAtMs: msg.CreatedAtMs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 不满足延迟条件则不存储直接投递
	if msg.NoDelay() {
		return id, p.delivery.DeliverImmediately(topic, id, message)
	}

	if store.SyncState != 0 {
		return 0, errors.New("can not write when sync data")
	}

	return id, p.store.Add(ctx, topic, id, message)
}

func (p *producer) Cancel(ctx context.Context, topic string, id uint64) error {
	return p.store.Delete(ctx, topic, id)
}
