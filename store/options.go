package store

import (
	redigo "github.com/garyburd/redigo/redis"
	"gorm.io/gorm"
)

type Options struct {
	Key      string //
	DB       *gorm.DB
	Pool     *redigo.Pool
	PerCount int
}

type Option func(*Options)
