package locker

import "github.com/garyburd/redigo/redis"

type RedisLockerOptions struct {
	Redis *redis.Pool
}

type RedisLockerOption func(options *RedisLockerOptions)
