package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	counterNewTag prometheus.Counter
)

func initPrometheus(env envConfig, mux *http.ServeMux) {

	counterNewTag = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "tag_count",
		Help:      "Count how many new tags have been found",
		Namespace: env.MetricsNamespace,
		Subsystem: env.MetricsSubsystem,
	})

	mux.Handle(env.MetricsPath, promhttp.Handler())
}
