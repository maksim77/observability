package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/juju/zaputil/zapctx"
	"gitlab.services.mts.ru/teta/golang-for-university/observability/internal/logger"
	"go.uber.org/zap"
	"moul.io/chizap"
)

func someFunc(ctx context.Context) {
	logger := zapctx.Logger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Hi from SomeFunc")
}

func main() {
	r := chi.NewRouter()

	logger, err := logger.GetLogger(true)
	if err != nil {
		log.Fatal(err)
	}

	logger = logger.With(zap.Any("someKey", "someValue"))

	r.Use(middleware.RequestID)
	r.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := zapctx.WithLogger(r.Context(), logger)
		someFunc(ctx)
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
