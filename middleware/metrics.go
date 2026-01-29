package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start).Seconds()

		// Record metrics
		requestsTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		requestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
		).Observe(duration)
	}
}
