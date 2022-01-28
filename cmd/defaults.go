package cmd

import (
	"dk/config"
	"dk/initial"
	"dk/messenger"
	"dk/store"
)

func setDefaults() {

	initial.DefDB = initial.InitGoOrm()
	initial.DefRedisPool = initial.InitRedis()
	initial.DefSyncProducer = initial.InitKafkaProducer()

	store.DefStore = store.NewBucketStore(func(opt *store.BucketStoreOptions) { opt.BucketCount = config.Cfg.BucketCount })
	messenger.DefDeliver = messenger.NewKafkaDelivery()
	messenger.DefProducer = messenger.NewProducer()

}
