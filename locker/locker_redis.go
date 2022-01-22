package locker

import (
	"context"

	"github.com/garyburd/redigo/redis"

	"kdqueue/initial"
)

type redisLocker struct {
	rc *redis.Pool
}

func NewRedisLocker(opts ...RedisLockerOption) Locker {

	opt := RedisLockerOptions{Redis: initial.DefRedisPool}

	for _, o := range opts {
		o(&opt)
	}
	return &redisLocker{rc: opt.Redis}
}

func (p *redisLocker) Lock(ctx context.Context, key, value string, expire int) error {

	conn := p.rc.Get()
	defer conn.Close()

	script := `
-- 不存在则直接 SETEX
if redis.call('EXISTS',KEYS[1])==0
then
	redis.call('SETEX',KEYS[1],ARGV[2],ARGV[1])
	return true
end
-- 存在则必须和 value 对应 且 EXPIRE 成功
return redis.call('GET',KEYS[1])==ARGV[1] and redis.call('EXPIRE',KEYS[1],ARGV[2])==1
`
	ok, err := redis.Bool(conn.Do("EVAL", script, 1, key, value, expire))
	if err != nil {
		return err
	}
	if !ok {
		return ErrLocked
	}
	return nil
}

func (p *redisLocker) UnLock(ctx context.Context, key, value string) error {

	conn := p.rc.Get()
	defer conn.Close()

	script := `
if redis.call('GET',KEYS[1])==ARGV[1] 
then
	redis.call('DEL',KEYS[1])
end
`
	_, err := conn.Do("EVAL", script, 1, key, value)
	return err
}
