package initial

import (
	"github.com/Shopify/sarama"

	"dk/config"
)

var DefSyncProducer sarama.SyncProducer

func InitKafkaProducer(opts ...KafkaOption) sarama.SyncProducer {

	opt := KafkaOptions{KafkaHost: config.Cfg.KafkaHost}

	for _, o := range opts {
		o(&opt)
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(opt.KafkaHost, cfg)
	if err != nil {
		panic(err)
	}
	return producer
}

type KafkaOptions struct {
	KafkaHost []string
}

type KafkaOption func(*KafkaOptions)
