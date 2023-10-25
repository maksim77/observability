package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.services.mts.ru/teta/golang-for-university/observability/internal/logger"
	"go.uber.org/zap"
)

func main() {
	r := chi.NewRouter()

	logger, err := logger.GetLogger(true)
	if err != nil {
		log.Fatal(err)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
