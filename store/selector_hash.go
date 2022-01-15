package store

// 使用消息 id hash

import (
	"hash/crc32"
	"strconv"

	"kdqueue/share/xid"
)

var defaultId = strconv.FormatUint(xid.Get(), 10)

// NewSelector returns an initalised round robin selector
func NewSelector(opts ...SelectorOption) Selector {
	return new(hashSelector)
}

type hashSelector struct{}

func (r *hashSelector) Select(stores []Store, opts ...SelectorOption) (Next, error) {

	if len(stores) == 0 {
		return nil, ErrNoneAvailable
	}

	opt := SelectorOptions{ID: defaultId} // default id
	for _, o := range opts {
		o(&opt)
	}

	return func() Store {
		cs := crc32.ChecksumIEEE([]byte(opt.ID))
		return stores[cs%uint32(len(stores))]
	}, nil
}

func (r *hashSelector) Record(addr Store, err error) error { return nil }

func (r *hashSelector) Reset() error { return nil }

func (r *hashSelector) String() string {
	return "hashSelector"
}
