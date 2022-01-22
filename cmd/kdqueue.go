package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"kdqueue/ha"
	"kdqueue/messenger"
	"kdqueue/restful"
	"kdqueue/share/health"
	"kdqueue/share/metric"
	"kdqueue/share/pprof"
	"kdqueue/store"
)

func QueueCommand() *cli.Command {

	return &cli.Command{
		Name:  `start`,
		Usage: `start the queue`,
		Action: func(ctx *cli.Context) error {

			setDefaults()

			health.RunOnPort()
			pprof.RunOnPort()
			metric.RunOnPort()

			h := ha.NewHA()
			err := errors.New("must master")
			for err != nil {
				err = h.MushMaster(ctx.Context)
				logrus.WithError(err).Warnln("waiting to become master")
				<-time.After(time.Second * 3)
			}
			logrus.Warnln("master node")

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
