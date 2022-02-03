package cmd

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"dk/ha"
	"dk/messenger"
	"dk/restful"
	"dk/share/health"
	"dk/share/metric"
	"dk/share/pprof"
	"dk/store"
)

func QueueCommand() *cli.Command {

	return &cli.Command{
		Name:  `start`,
		Usage: `start the queue`,
		Action: func(ctx *cli.Context) error {

			setDefaults()

			health.RunOnPort()
			pprof.RunOnPort()

			h := ha.NewHA()

			for {
				if err := h.MushMaster(ctx.Context); err == nil {
					break
				}
				logrus.Warnln("waiting to become master")
				<-time.After(time.Second * 3)
			}
			logrus.Warnln("master node")

			metric.RunOnPort()

			c, cancel := context.WithCancel(ctx.Context)
			defer cancel()

			errCh := make(chan error, 3)

			go func() {
				errCh <- store.NewDatadog().Sync(c)
			}()

			go func() {
				errCh <- restful.Run(c)
			}()

			go func() {
				errCh <- messenger.DefDeliver.DoWork(c)
			}()

			return <-errCh
		},
	}
}
