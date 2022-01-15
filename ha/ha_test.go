package ha

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "kdqueue/xtesting"
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
