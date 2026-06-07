package main

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const name = "go.opentelemetry.io/contrib/examples/dice"

var (
	tracer  = otel.Tracer(name)
	meter   = otel.Meter(name)
	logger  = otelslog.NewLogger(name)
	rollCnt metric.Int64Counter
)

func init() {
	var err error
	rollCnt, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))

	if err != nil {
		panic(err)
	}
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	roll := 1 + rand.Intn(6)

	var msg string
	var player string
	if player = r.PathValue("player"); player != "" {
		msg = player + " is rolling the dice"
	} else {
		msg = "Anonymous is rolling the dice"
	}
	logger.InfoContext(ctx, msg, "result", roll)

	// set attributes to span
	if player != "" {
		span.SetAttributes(attribute.String("player.name", player))
	}
	span.SetAttributes(attribute.Int("dice.sides", 6))

	rollValueAttr := attribute.Int("roll.value", roll)

	// add event to span
	span.AddEvent("dice rolled", trace.WithAttributes(
		rollValueAttr,
	))

	// add attributes to metric
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		// set error and status to span
		span.RecordError(err)
		span.SetStatus(codes.Error, "write response failed")
		logger.ErrorContext(ctx, "Write failed", "error", err)
	}
}
