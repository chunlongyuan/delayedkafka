package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"dk/cmd"
	"dk/config"
)

func init() {

	if len(os.Getenv("ENV")) == 0 {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}
		log.Trace("dot env loaded")
	}

	err := env.Parse(&config.Cfg)
	if err != nil {
		panic(err)
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	if level, err := log.ParseLevel(config.Cfg.LogLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infof("got config:%#v", config.Cfg)
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	app := cli.NewApp()
	app.Name = "dk"
	app.Usage = "A kafka and mysql backed priority queue for scheduling delayed events"
	app.Commands = []*cli.Command{
		cmd.QueueCommand(),
		cmd.SyncTableForTest(),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		err := app.RunContext(ctx, os.Args)
		if err != nil {
			log.WithError(err).Error("run stop")
		}
	}()

	log.Errorln("application start")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	select {
	case <-sig:
		log.Errorln("receive stop signal")
		cancel()
	case <-done:
		log.Errorln("receive done")
	}

	log.Errorln("before done")
	<-done
	log.Errorln("after done")

	<-time.After(1 * time.Second)
}
