package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const name = "lucky-checker"

var tracer = otel.Tracer(name)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/lucky/{value}", luckyHandler)

	srv := &http.Server{
		Addr:         ":8081",
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      otelhttp.NewHandler(mux, "/"),
	}

	srvErr := make(chan error, 1)
	go func() {
		log.Println("Lucky service running on :8081")
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	return srv.Shutdown(context.Background())
}

func luckyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "check-lucky")
	defer span.End()

	raw := r.PathValue("value")
	value, err := strconv.Atoi(raw)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid value")
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	lucky := isLucky(ctx, value)

	span.SetAttributes(
		attribute.Int("dice.value", value),
		attribute.Bool("lucky", lucky),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"value": value,
		"lucky": lucky,
	})
}

func isLucky(ctx context.Context, value int) bool {
	_, span := tracer.Start(ctx, "isLucky")
	defer span.End()

	lucky := value == 6
	span.SetAttributes(attribute.Bool("lucky", lucky))
	return lucky
}
