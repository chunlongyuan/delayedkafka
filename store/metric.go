package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	delayedMessagesTotal *prometheus.GaugeVec
)

func init() {
	delayedMessagesTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_delayed_messages_total", "delayedqueue"),
		Help: "Delayed message count.",
	}, []string{"bucket"})
	prometheus.MustRegister(delayedMessagesTotal)
}

func metricMessageTotal(bucket string, total int) {
	delayedMessagesTotal.WithLabelValues(bucket).Set(float64(total))
}
