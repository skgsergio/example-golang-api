package demo

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// RegisterHandlers registers all handlers for this module
func RegisterHandlers() {
	log.Info().Msg("Registering demo module handlers...")
	http.HandleFunc("/demo", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_, err := fmt.Fprint(w, `{"patatas":1337,"type":"fritas"}`)
	if err != nil {
		log.Error().Err(err).Msg("Response write error.")
	}
}
