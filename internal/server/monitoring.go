package server

import "github.com/prometheus/client_golang/prometheus"

var (
	sessionsGauge           prometheus.Gauge
	outgoingMessagesCounter *prometheus.CounterVec
)

func init() {
	sessionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "rms",
		Name:      "ws_sessions",
		Help:      "Count of websocket clients",
	})

	outgoingMessagesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "rms",
		Name:        "outgoing_messages_count",
		Help:        "Total count of messages from device",
		ConstLabels: nil,
	}, []string{"device"})
}
