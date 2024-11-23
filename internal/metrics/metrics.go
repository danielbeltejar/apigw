package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	RequestsTotal           = prometheus.NewCounter(prometheus.CounterOpts{Name: "http_requests_total", Help: "Total requests received"})
	SuccessfulRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{Name: "http_requests_success_total", Help: "Successful requests"})
	MethodNotAllowedTotal   = prometheus.NewCounter(prometheus.CounterOpts{Name: "http_requests_method_not_allowed_total", Help: "Method not allowed requests"})
	NotFoundTotal           = prometheus.NewCounter(prometheus.CounterOpts{Name: "http_requests_not_found_total", Help: "Requests not found"})
)

func InitMetrics() {
	prometheus.MustRegister(RequestsTotal, SuccessfulRequestsTotal, MethodNotAllowedTotal, NotFoundTotal)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
