package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// httpResponseCount stores the count of responses of each endpoint and HTTP code pair
	httpResponseCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Subsystem: "response",
			Name:      "count",
			Help:      "HTTP endpoint response count",
		},
		[]string{"endpoint", "code"},
	)

	// httpResponseTime stores the response time of each endpoint
	httpResponseTime = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "http",
			Subsystem:  "response",
			Name:       "time",
			Help:       "HTTP endpoint response time in milliseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"endpoint"},
	)
)

// metricsResponseWriter is a custom http.ResponseWritter that stores the response
// status code and content length.
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

// WriteHeader calls the original ResponseWriter.WriteHeader() and stores the status code.
func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.ResponseWriter.WriteHeader(statusCode)
	mrw.statusCode = statusCode
}

// Write calls the original ResponseWriter.Write() and stores the number of written bytes.
func (mrw *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := mrw.ResponseWriter.Write(b)
	mrw.contentLength += n
	return n, err
}

// LoggerAndMetrics is a custom net/http middleware for detailed request logging
// and storing metrics.
func LoggerAndMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Encapsulate the original response writer with our own that stores metrics
		mrw := &metricsResponseWriter{ResponseWriter: w}

		// Perform the request timming it
		start := time.Now()
		next.ServeHTTP(mrw, r)
		duration := time.Since(start)

		if mrw.statusCode == 0 {
			mrw.statusCode = 200
		}

		// Log request with the appropiate log level
		level := zerolog.InfoLevel
		if mrw.statusCode >= 500 {
			level = zerolog.ErrorLevel
		} else if mrw.statusCode >= 400 {
			level = zerolog.WarnLevel
		}

		log.WithLevel(level).
			Str("remote_addr", strings.Split(r.RemoteAddr, ":")[0]).
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Int("status_code", mrw.statusCode).
			Int("content_length", mrw.contentLength).
			Dur("response_time", duration).
			Msg("Request")

		// Store request metrics, ignore 404 errors
		if mrw.statusCode != 404 {
			httpResponseCount.WithLabelValues(r.URL.Path, fmt.Sprintf("%d", mrw.statusCode)).Inc()
			httpResponseTime.WithLabelValues(r.URL.Path).Observe(float64(duration.Milliseconds()))
		}
	})
}
