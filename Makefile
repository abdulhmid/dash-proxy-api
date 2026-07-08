.PHONY: build run test clean migrate

APP_NAME = api-source-proxy

build:
	go build -o bin/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

clean:
	rm -rf bin/

migrate:
	psql -h localhost -U postgres -d api_source_proxy -f migrations/postgres/001_create_users.up.sql
	psql -h localhost -U postgres -d api_source_proxy -f migrations/postgres/002_create_api_keys.up.sql
	psql -h localhost -U postgres -d api_source_proxy -f migrations/postgres/003_create_api_sources.up.sql

docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

.PHONY: lint
lint:
	golangci-lint run ./...
