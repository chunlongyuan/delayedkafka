package ha

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"dk/config"
	"dk/share/ip"
	_ "dk/xtesting"
)

func TestHa(t *testing.T) {

	var (
		ctx = context.Background()
	)

	// 同一个 node 总是成功
	assert.Nil(t, NewHA(func(opt *Options) { opt.NodeId = "1" }).MushMaster(ctx))
	assert.Nil(t, NewHA(func(opt *Options) { opt.NodeId = "1" }).MushMaster(ctx))
	// 此时其他 node 要失败
	assert.Equal(t, ErrNotMaster, NewHA(func(opt *Options) { opt.NodeId = "2" }).MushMaster(ctx))
}

func Test_genNodeId(t *testing.T) {

	// got privateIp when not customize
	assert.True(t, func() bool {
		id := genNodeId()
		return len(id) > 0 && id == ip.PrivateIPv4()
	}())

	// got customize
	config.Cfg.NodeId = "helloworld"
	assert.Equal(t, genNodeId(), "helloworld")
}
