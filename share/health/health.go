package health

import (
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"

	"kdqueue/config"
)

func RunOnPort() {
	addr := net.JoinHostPort("", config.Cfg.HealthPort)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Errorf("run health err(%v)", err)
		}
	}()
}
