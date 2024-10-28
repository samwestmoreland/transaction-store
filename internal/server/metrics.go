package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	requestDuration prometheus.Histogram
	requestSuccess  prometheus.Counter
	requestErrors   prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		requestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "transaction_request_duration_seconds",
			Help: "Time taken to process transaction requests",
			// You might want to tune these buckets based on expected latencies
			Buckets: prometheus.DefBuckets,
		}),
		requestSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "transaction_requests_success_total",
			Help: "Total number of successful transaction requests",
		}),
		requestErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "transaction_requests_errors_total",
			Help: "Total number of failed transaction requests",
		}),
	}
}
