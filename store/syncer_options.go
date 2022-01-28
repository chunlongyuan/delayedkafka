package store

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"gorm.io/gorm"
)

type SyncerOptions struct {
	Store Store
	DB    *gorm.DB
	Redis *redis.Pool
	//
	MonitordkSeconds int
	//
	MonitorInterval time.Duration
}

type SyncerOption func(options *SyncerOptions)
