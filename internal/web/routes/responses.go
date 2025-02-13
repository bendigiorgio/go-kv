package routes

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	jsonPay, err := json.Marshal(payload)

	if err != nil {
		log.Error().Stack().Err(err).Msg("Error when marshaling JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(jsonPay)
}
