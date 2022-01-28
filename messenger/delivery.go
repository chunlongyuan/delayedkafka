package messenger

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"

	"dk/initial"
	"dk/store"
)

var (
	DefDeliver Delivery
)

// Delivery 负责消费队列里的消息并投递到 kafka 中
type Delivery interface {
	DoWork(ctx context.Context) error
	DeliverImmediately(topic string, id uint64, msg store.Message) error
}

// KafkaMessage 定义投递的消息结构
type KafkaMessage struct {
	Id          string    `json:"id"`
	DelaySecond int64     `json:"delay_second"`  // 延迟多少毫秒
	CreatedAtMs int64     `json:"created_at_ms"` // 毫秒时间戳
	Body        string    `json:"body"`          // 元数据
	IssuedAt    time.Time `json:"issued_at"`
}

type kafkaDelivery struct {
	store    store.Store
	producer sarama.SyncProducer
}

func NewKafkaDelivery(opts ...Option) Delivery {

	opt := Options{Store: store.DefStore, Producer: initial.DefSyncProducer}

	for _, o := range opts {
		o(&opt)
	}

	return &kafkaDelivery{store: opt.Store, producer: opt.Producer}
}

// DoWork postman do his work
func (p *kafkaDelivery) DoWork(ctx context.Context) error {

	return p.store.FetchDelayMessage(ctx, func(topic string, id uint64, msg store.Message) error {
		// deliver these message
		return p.DeliverImmediately(topic, id, store.Message{
			ID:          id,
			Topic:       topic,
			TTRms:       time.Unix(0, msg.CreatedAtMs*1e6).Add(time.Duration(msg.DelayMs)*time.Millisecond).UnixNano() / 1e6,
			DelayMs:     msg.DelayMs,
			Body:        msg.Body,
			CreatedAtMs: msg.CreatedAtMs,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	})
}

func (p *kafkaDelivery) DeliverImmediately(topic string, id uint64, msg store.Message) error {

	postmanMessage := KafkaMessage{
		Id:          strconv.FormatUint(id, 10),
		DelaySecond: msg.DelayMs / 1e3,
		Body:        msg.Body,
		CreatedAtMs: msg.CreatedAtMs,
		IssuedAt:    time.Now(),
	}

	body, err := json.Marshal(&postmanMessage)
	if err != nil {
		logrus.WithError(err).WithField("postman_message", postmanMessage).Errorln("marshal err")
		return err
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(body),
	})
	return err
}

type Options struct {
	Store    store.Store
	Producer sarama.SyncProducer
}

type Option func(options *Options)
