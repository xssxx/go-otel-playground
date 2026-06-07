.PHONY: build run test tidy docker-build up down logs restart clean

APP_NAME := server
BINARY := ./$(APP_NAME)

build:
	go build -o $(APP_NAME) .

run:
	go run .

test:
	go test ./...

tidy:
	go mod tidy

docker-build:
	docker build -t otel-app .

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose restart app

clean:
	docker compose down -v --remove-orphans
	rm -f $(APP_NAME)
