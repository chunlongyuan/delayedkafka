package initial

import (
	"time"

	redigo "github.com/garyburd/redigo/redis"

	"kdqueue/config"
)

var (
	DefRedisPool *redigo.Pool
)

func InitRedis(opts ...RedisOption) *redigo.Pool {

	cfg := config.Cfg

	opt := RedisOptions{Host: cfg.RedisCacheHost + ":" + cfg.RedisCachePort}
	for _, o := range opts {
		o(&opt)
	}

	pool := &redigo.Pool{
		MaxIdle:     500,
		MaxActive:   1200,
		IdleTimeout: time.Second * 30,
		Wait:        true,
		Dial: func() (redigo.Conn, error) {
			return redigo.Dial("tcp", opt.Host)
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return pool
}

type RedisOptions struct {
	Host string
}
type RedisOption func(options *RedisOptions)
