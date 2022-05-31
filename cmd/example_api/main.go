package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/skgsergio/example-golang-api/internal/codes"
	"github.com/skgsergio/example-golang-api/internal/demo"

	"github.com/skgsergio/example-golang-api/lib/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	listen             = flag.String("listen", ":8000", "API listen address")
	trustedProxyCIDR   = flag.String("trusted-proxy-cidr", "", "proxy CIDR for trusting X-Forwarded-For header")
	insecureTrustProxy = flag.Bool("insecure-trust-proxy", false, "always trust X-Forwarded-For")
	debug              = flag.Bool("debug", false, "enable debug log level")
	pretty             = flag.Bool("pretty", false, "enable pretty logging (human-friendly)")
)

func main() {
	flag.Parse()

	// Set logger preferences
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	log.Info().
		Str("listen", *listen).
		Str("log_level", zerolog.GlobalLevel().String()).
		Msg("Starting API...")

	// Register prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Dumb health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "Healthy!\n")
		if err != nil {
			log.Error().Err(err).Msg("Response write error.")
		}
	})

	// Register app handlers
	demo.RegisterHandlers()
	codes.RegisterHandlers()

	// Run server with custom LoggerAndMetrics middleware
	err := http.ListenAndServe(*listen, middleware.LoggerAndMetrics(http.DefaultServeMux, *trustedProxyCIDR, *insecureTrustProxy))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed serving app")
	}
}
