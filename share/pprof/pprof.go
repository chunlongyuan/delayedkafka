package pprof

import (
	"net"
	"net/http"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"

	"dk/config"
)

func RunOnPort() {
	go func() {
		if err := http.ListenAndServe(net.JoinHostPort("", config.Cfg.PProfPort), nil); err != nil {
			log.WithError(err).Error("start pprof err")
		}
	}()
}
