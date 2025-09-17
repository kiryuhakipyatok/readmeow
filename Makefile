-include .env
export
version=
name=

all: build test

build:
	@echo "Building..."
	@go build -o main.exe cmd/app/main.go

run:
	@go run cmd/app/main.go

docker-run-infra:
	@docker-compose up -d postgres redis elasticsearch

docker-run-elk:
	@docker-compose up -d kibana filebeat logstash

docker-run-app:
	docker-compose up -d readmeow --build

docker-run-monitoring:
	@docker-compose up -d  prometheus grafana

docker-app-logs:
	@docker-compose logs -f readmeow

docker-run-all: docker-run-infra docker-run-elk docker-run-app docker-run-monitoring docker-app-logs

docker-down:
	@docker-compose down

docker-build:
	@docker-compose build --no-cache

test:
	@echo "Testing..."
	@go test ./... -v

clean:
	@echo "Cleaning..."
	@rm -f main.exe

docker-migrate-up:
	@docker-compose run --rm migrate up

docker-migrate-down:
	@docker-compose run --rm migrate down


create:
	@goose -dir=$(MIGRATIONS_PATH) create $(NAME) sql

down-to:
	@docker-compose run --rm migrate down-to $(VERSION)

status:
	@docker-compose run --rm migrate status

reset:
	@docker-compose run --rm migrate reset