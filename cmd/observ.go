package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/juju/zaputil/zapctx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	"gitlab.services.mts.ru/teta/golang-for-university/observability/internal/logger"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"moul.io/chizap"
)

const (
	DSN         = "https://9cb3a25fd4b9a73e6630f86c06a10208@o599355.ingest.sentry.io/4506109917134848"
	TRACER_NAME = "demo_service"
)

func someFunc(ctx context.Context) {
	_, span := tracer.Start(ctx, "someFunc")
	defer span.End()

	logger := zapctx.Logger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Hi from SomeFunc")
}

func someFuncWithError(ctx context.Context) error {
	logger := zapctx.Logger(ctx)
	logger.Info("start someFuncWithError")
	return errors.New("Oops... one more error")
}

var tracer = otel.Tracer(TRACER_NAME)

func main() {
	shutdown := initProvider()
	defer shutdown()

	r := chi.NewRouter()
	r.Use(otelchi.Middleware("main", otelchi.WithChiRoutes(r)))

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

	counter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "teta", Name: "testcounter", Help: "Main endpoint request counter",
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		newCtx, span := tracer.Start(r.Context(), "/")
		defer span.End()

		counter.Inc()
		ctx := zapctx.WithLogger(newCtx, logger)
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

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9000", nil) //nolint:errcheck

	logger.Info("Server started")
	http.ListenAndServe(":8080", r) //nolint:errcheck
}
