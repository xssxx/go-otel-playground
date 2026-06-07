package main

import (
	"context"
	"database/sql"
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

func rolldice(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "roll")
		defer span.End()

		var player string
		if player = r.PathValue("player"); player != "" {
			logger.InfoContext(ctx, player+" is rolling the dice")
		} else {
			logger.InfoContext(ctx, "Anonymous is rolling the dice")
		}

		// set attributes to span
		if player != "" {
			span.SetAttributes(attribute.String("player.name", player))
		}
		span.SetAttributes(attribute.Int("dice.sides", 6))

		roll, err := computeRoll(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "compute roll failed")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		rollValueAttr := attribute.Int("roll.value", roll)
		span.AddEvent("dice rolled", trace.WithAttributes(rollValueAttr))
		rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

		if err := saveResult(ctx, db, player, roll); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "save result failed")
			logger.ErrorContext(ctx, "Save failed", "error", err)
		}

		resp := strconv.Itoa(roll) + "\n"
		if _, err := io.WriteString(w, resp); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "write response failed")
			logger.ErrorContext(ctx, "Write failed", "error", err)
		}
	}
}

func computeRoll(ctx context.Context) (int, error) {
	_, span := tracer.Start(ctx, "computeRoll")
	defer span.End()

	roll := 1 + rand.Intn(6)
	span.SetAttributes(attribute.Int("roll.value", roll))
	return roll, nil
}

func saveResult(ctx context.Context, db *sql.DB, player string, roll int) error {
	_, span := tracer.Start(ctx, "saveResult")
	defer span.End()

	if player == "" {
		player = "Anonymous"
	}

	query := "INSERT INTO rolls (player, result) VALUES (?, ?)"
	span.SetAttributes(
		attribute.String("db.system", "sqlite"),
		attribute.String("db.statement", query),
		attribute.String("player.name", player),
		attribute.Int("roll.value", roll),
	)

	_, err := db.ExecContext(ctx, query, player, roll)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db insert failed")
		return err
	}

	return nil
}
