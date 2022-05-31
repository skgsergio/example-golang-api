package codes

import (
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// RegisterHandlers registers all handlers for this module
func RegisterHandlers() {
	log.Info().Msg("Registering codes module handlers...")

	for code := 100; code < 600; code++ {
		if http.StatusText(code) != "" {
			http.HandleFunc(fmt.Sprintf("/codes/%d", code), reply)
		}
	}
}

func reply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse code
	code, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		log.Error().Err(err).Msg("Error parsing code")
		code = http.StatusInternalServerError
	}

	// Add some logging for errors and some artificial latency for testing metrics
	if code >= 500 {
		time.Sleep(time.Duration(rand.Intn(750)) * time.Millisecond)
		log.Error().Int("code", code).Msg("Server error")
	} else if code >= 400 {
		time.Sleep(time.Duration(rand.Intn(250)) * time.Millisecond)
		log.Warn().Int("code", code).Msg("Client error")
	}

	// Send the response
	w.WriteHeader(code)
	_, err = fmt.Fprintf(w, `{"sent_code": %d, "meaning": "%s"}`, code, http.StatusText(code))
	if err != nil {
		log.Error().Err(err).Msg("Response write error.")
	}
}
