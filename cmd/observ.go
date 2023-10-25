package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.services.mts.ru/teta/golang-for-university/observability/internal/logger"
	"go.uber.org/zap"
	"moul.io/chizap"
)

func main() {
	r := chi.NewRouter()

	logger, err := logger.GetLogger(false)
	if err != nil {
		log.Fatal(err)
	}

	r.Use(middleware.RequestID)
	r.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
