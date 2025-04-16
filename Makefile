
export $(shell grep -v '^#' .env | xargs)

migrate:
	bash scripts/migrate.sh

run:
	go run cmd/api/main.go

swag:
	swag init -g cmd/api/main.go -o docs

build:
	go build -o bin/people-enricher cmd/api/main.go

lint: 
	golangci-lint run ./...