package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/juju/zaputil/zapctx"
	"gitlab.services.mts.ru/teta/golang-for-university/observability/internal/logger"
	"go.uber.org/zap"
	"moul.io/chizap"
)

func sentryInit() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://9cb3a25fd4b9a73e6630f86c06a10208@o599355.ingest.sentry.io/4506109917134848",
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
}

func someFunc(ctx context.Context) {
	logger := zapctx.Logger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Hi from SomeFunc")
}

func main() {
	sentryInit()
	defer sentry.Flush(time.Second * 2)

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
			sentry.CaptureException(err)
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		err := errors.New("Oops! Some error")
		sentry.CaptureException(err)
		_, err = w.Write([]byte("Oops!"))
		w.WriteHeader(http.StatusInternalServerError)
		if err != nil {
			sentry.CaptureException(err)
			logger.Error("Error writing response", zap.Error(err))
		}
	})

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
