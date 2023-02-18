package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	sessionsGauge           prometheus.Gauge
	outgoingMessagesCounter *prometheus.CounterVec
)

func init() {
	sessionsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "rms",
		Subsystem: "bot_server",
		Name:      "sessions",
		Help:      "Count of websocket clients",
	})

	outgoingMessagesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "rms",
		Subsystem:   "bot_server",
		Name:        "outgoing_messages_count",
		Help:        "Total count of messages from device",
		ConstLabels: nil,
	}, []string{"device"})
}
