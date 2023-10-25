package main

import (
	"context"
	"errors"
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

const (
	DSN = "https://9cb3a25fd4b9a73e6630f86c06a10208@o599355.ingest.sentry.io/4506109917134848"
)

func someFunc(ctx context.Context) {
	logger := zapctx.Logger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Hi from SomeFunc")
}

func someFuncWithError(ctx context.Context) error {
	logger := zapctx.Logger(ctx)
	logger.Info("start someFuncWithError")
	return errors.New("Oops... one more error")
}

func main() {
	r := chi.NewRouter()

	logger, err := logger.GetLogger(false, DSN, "production")
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

	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		err := someFuncWithError(r.Context())
		if err != nil {
			logger.Error("Error", zap.Error(err))
		}

		_, err = w.Write([]byte("Oops!"))
		w.WriteHeader(http.StatusInternalServerError)
		if err != nil {
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
