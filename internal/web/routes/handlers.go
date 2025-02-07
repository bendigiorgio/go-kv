package routes

import (
	"log"
	"net/http"

	"github.com/bendigiorgio/go-kv/internal/api/api_errors"
	"github.com/bendigiorgio/go-kv/internal/engine"
)

type CustomHandler func(w http.ResponseWriter, request *http.Request, engine *engine.Engine) error

func NewHandler(customHandler CustomHandler, engine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := customHandler(w, r, engine)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			if clientErr, ok := err.(*api_errors.ClientErr); ok {
				respondWithJSON(w, clientErr.HttpCode, clientErr)
			} else {
				respondWithJSON(w, http.StatusInternalServerError,
					api_errors.InternalErr{
						HttpCode: http.StatusInternalServerError,
						Message:  "internal server error",
					},
				)
			}
		}
	}
}

type CustomHandlerNoEngine func(w http.ResponseWriter, request *http.Request) error

func NewHandlerNoEngine(customHandler CustomHandlerNoEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := customHandler(w, r)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			if clientErr, ok := err.(*api_errors.ClientErr); ok {
				respondWithJSON(w, clientErr.HttpCode, clientErr)
			} else {
				respondWithJSON(w, http.StatusInternalServerError,
					api_errors.InternalErr{
						HttpCode: http.StatusInternalServerError,
						Message:  "internal server error",
					},
				)
			}
		}
	}
}
