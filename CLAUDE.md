# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run          # Run the app locally (requires an OTLP collector on localhost:4317)
make build        # Compile binary to ./server
make test         # Run all tests
make tidy         # Sync go.mod / go.sum

make up           # Build and start app + Jaeger via Docker Compose (detached)
make down         # Stop containers
make logs         # Tail container logs
make restart      # Restart only the app container
make clean        # Remove containers, volumes, and the local binary
```

## Architecture

This is a minimal Go HTTP service that demonstrates the OpenTelemetry Go SDK.

**Signal pipeline (`otel.go`)** — `setupOTelSDK` wires up three OTel signals:
- **Traces** — exported via OTLP/gRPC to the endpoint in `OTEL_EXPORTER_OTLP_ENDPOINT` (default `localhost:4317`). In the Docker Compose setup this points to Jaeger.
- **Metrics** — exported to stdout (pretty-printed) on a 3-second interval via `newMeterProvider` (the provider is created but **not registered** in `setupOTelSDK` — a known gap).
- **Logs** — exported to stdout (pretty-printed) via the OTel log bridge.

**HTTP layer (`main.go`)** — single `net/http` mux wrapped with `otelhttp.NewHandler` for automatic span creation on every request.

**Business logic (`rolldice.go`)** — `/rolldice` and `/rolldice/{player}` endpoints. Each request creates a child span (`roll`), increments the `dice.rolls` Int64Counter metric with a `roll.value` attribute, and emits a structured log line via `otelslog`.

**Observability infra (`docker-compose.yml`)** — Jaeger all-in-one receives OTLP gRPC on port 4317; its UI is at `http://localhost:16686`. The app container sets `OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317`.

## Environment

Copy `.env.example` to `.env` if overrides are needed. The only runtime env var the app reads is `OTEL_EXPORTER_OTLP_ENDPOINT`.
