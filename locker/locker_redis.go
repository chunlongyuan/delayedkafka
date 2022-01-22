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
if redis.call('EXISTS',KEYS[1])==0
then
	redis.call('SETEX',KEYS[1],ARGV[2],ARGV[1])
	return 1
end

if redis.call('GET',KEYS[1])==ARGV[1]
then
	redis.call('EXPIRE',KEYS[1],ARGV[2])
	return 1
end
return 0
`
	code, err := redis.Int(conn.Do("EVAL", script, 1, key, value, expire))
	if err != nil {
		return err
	}
	if code != 1 {
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
