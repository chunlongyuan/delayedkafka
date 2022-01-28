package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	minWaitDuration = time.Second
)

var (
	DefStore Store
)

// 负责将数据放到不同的桶里
type bucketStore struct {
	stores       []Store
	selector     Selector
	waitDuration time.Duration
}

func NewBucketStore(opts ...BucketStoreOption) Store {

	// 默认一个桶
	opt := BucketStoreOptions{BucketCount: 1, Selector: NewSelector(), WaitDuration: minWaitDuration} // 默认 hash 分桶

	for _, o := range opts {
		o(&opt)
	}

	if opt.WaitDuration < minWaitDuration {
		panic(fmt.Sprintf(`idle duration must greater than %v`, minWaitDuration))
	}

	bs := bucketStore{selector: opt.Selector, waitDuration: opt.WaitDuration}

	newStoreFun := func(i int) {
		bs.stores = append(bs.stores, NewStore(func(opt *Options) {
			opt.Key = fmt.Sprintf(`%02d`, i)
		}))
	}

	for i := 0; i < opt.BucketCount; i++ {
		newStoreFun(i + 1)
	}

	return &bs
}

func (p *bucketStore) Add(ctx context.Context, topic string, id uint64, msg Message) error {

	// 根据算法取一个 store
	storeSelector, err := p.selector.Select(p.stores, func(opt *SelectorOptions) {
		opt.ID = strconv.FormatUint(id, 10)
	})
	if err != nil {
		return err
	}

	s := storeSelector()

	p.selector.Record(s, err)

	return s.Add(ctx, topic, id, msg)
}

func (p *bucketStore) Delete(ctx context.Context, topic string, id uint64) error {

	// 根据算法取一个 store
	storeSelector, err := p.selector.Select(p.stores, func(opt *SelectorOptions) {
		opt.ID = strconv.FormatUint(id, 10)
	})
	if err != nil {
		return err
	}

	s := storeSelector()

	p.selector.Record(s, err)

	return s.Delete(ctx, topic, id)
}

func (p *bucketStore) FetchDelayMessage(ctx context.Context, handle func(topic string, id uint64, msg Message) error) error {

	var eg errgroup.Group

	fetchFun := func(s Store) func() error {

		return func() error {

			for errTimes := 0; ; {

				select {
				case <-ctx.Done():
					return context.DeadlineExceeded
				default:
				}

				err := s.FetchDelayMessage(ctx, handle)

				if err == nil { // 可能还有数据需要处理
					errTimes = 0
					continue
				}

				if errors.Is(err, ErrNoData) { // 没数据了
					errTimes = 0
					<-time.After(p.waitDuration) // 常规等待
					continue
				}

				// 上层处理失败, 异常等待指数退避
				waitDuration := p.waitDuration << errTimes
				logrus.WithError(err).WithField("errTimes", errTimes).WithField("waitDuration", waitDuration).Errorln("handle delay message err")

				<-time.After(waitDuration)
			}
		}
	}

	// 每个 bucket 都有一个 routine 在处理
	for _, ss := range p.stores {
		eg.Go(fetchFun(ss))
	}

	return eg.Wait()
}
