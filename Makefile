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
	@docker-compose up postgres redis es --build

docker-run-app:
	@docker-compose up app --build

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

up:
	@goose -dir=$(MIGRATIONS_PATH) up

down:
	@goose -dir=$(MIGRATIONS_PATH) down

down-to:
	@goose -dir=$(MIGRATIONS_PATH) down-to $(VERSION)

status:
	@goose -dir=$(MIGRATIONS_PATH) status

reset:
	@goose -dir=$(MIGRATIONS_PATH) reset