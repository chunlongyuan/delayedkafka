package metric

import (
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"dk/config"
)

var (
	namespace = config.Cfg.QueueKeyword
)

func RunOnPort() {
	go func() {
		addr := net.JoinHostPort("", config.Cfg.PrometheusPort)
		http.Handle("/metrics", promhttp.Handler())
		log.Errorln(http.ListenAndServe(addr, nil)) //start http server
	}()
}
