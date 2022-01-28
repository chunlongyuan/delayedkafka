package store

import (
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"gorm.io/gorm"
)

type Options struct {
	Key            string
	DB             *gorm.DB
	Pool           *redigo.Pool
	PerCount       int
	MetricInterval time.Duration
}

type Option func(*Options)
