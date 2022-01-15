package locker

import (
	"context"
	"fmt"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"

	"kdqueue/initial"
	_ "kdqueue/xtesting"
)

func TestRedisLocker_Lock(t *testing.T) {

	var (
		ctx    = context.Background()
		key    = "key"
		value  = "value"
		expire = 3
		rc     = initial.DefRedisPool
	)

	conn := rc.Get()
	defer conn.Close()

	locker := NewRedisLocker()

	// 多次设置都应该成功
	assert.Nil(t, locker.Lock(ctx, key, value, expire))
	// key 和 value 相同时更新 expire
	assert.Nil(t, locker.Lock(ctx, key, value, 10))
	num, err := redis.Int(conn.Do("ttl", key))
	assert.Nil(t, err)
	assert.Equal(t, 10, num)

	// key 相同但是 value 不同 lock 失败
	fmt.Println(locker.Lock(ctx, key, value+"-1", expire))
	assert.NotNil(t, locker.Lock(ctx, key, value+"-1", expire))
	assert.NotNil(t, locker.Lock(ctx, key, value+"-2", expire))

	// key 不同 lock 成功
	assert.Nil(t, locker.Lock(ctx, key+"-1", value, expire))

	// 取消成功
	assert.Nil(t, locker.UnLock(ctx, key, value))
	assert.Nil(t, locker.UnLock(ctx, key+"-1", value))
}
