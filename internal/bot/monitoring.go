package bot

import "github.com/prometheus/client_golang/prometheus"

var (
	incomingMessagesCounter *prometheus.CounterVec
)

func init() {
	incomingMessagesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "rms",
		Name:        "incoming_messages_count",
		Help:        "Total count of messages from user",
		ConstLabels: nil,
	}, []string{"user"})
}
