package store

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"gorm.io/gorm"

	"kdqueue/ha"
)

type SyncerOptions struct {
	HA    ha.HA
	Store Store
	DB    *gorm.DB
	Redis *redis.Pool
	//
	MonitorKDQueueSeconds int
	//
	MonitorInterval time.Duration
}

type SyncerOption func(options *SyncerOptions)
