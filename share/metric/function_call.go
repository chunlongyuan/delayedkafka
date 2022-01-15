package metric

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	functionCalls        *prometheus.CounterVec
	functionCallDuration *prometheus.GaugeVec
)

func init() {
	functionCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_operations_total", namespace),
		Help: "Total number of operations.",
	}, []string{"class", "function", "result"})
	prometheus.MustRegister(functionCalls)

	functionCallDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: fmt.Sprintf("%s_operate_duration_seconds", namespace),
		Help: "Total operation time.",
	}, []string{"class", "function"})
	prometheus.MustRegister(functionCallDuration)
}

func FunctionCall(startTime time.Time, class, function string) {
	functionCallDuration.WithLabelValues(class, function).Set(time.Since(startTime).Seconds())
	result := "ok"
	functionCalls.WithLabelValues(class, function, result).Inc()
}
