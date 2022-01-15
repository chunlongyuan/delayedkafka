package cmd

import (
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"kdqueue/messenger"
	"kdqueue/restful"
	"kdqueue/share/health"
	"kdqueue/share/metric"
	"kdqueue/share/pprof"
	"kdqueue/store"
)

func KQueueCommand() *cli.Command {

	return &cli.Command{
		Name:  `start`,
		Usage: `start the queue`,
		Action: func(ctx *cli.Context) error {

			setDefaults()

			health.RunOnPort()
			pprof.RunOnPort()
			metric.RunOnPort()

			var eg errgroup.Group

			eg.Go(func() error { // 负责守护 redis 不丢数据
				return store.NewSyncer().Sync(ctx.Context)
			})

			eg.Go(func() error { // 负责投递
				return messenger.DefDeliver.DoWork(ctx.Context)
			})

			eg.Go(func() error { // 提供操作消息的 restful api
				return restful.Run(ctx.Context)
			})

			return eg.Wait()
		},
	}
}
