# syntax=docker/dockerfile:1.6
FROM golang:1.25-bookworm AS builder

WORKDIR /usr/local/src

RUN apt-get update && apt-get install -y ca-certificates tzdata make

COPY go.mod go.sum ./
RUN go mod download 

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go install -tags='no_sqlite no_mysql no_clickhouse no_mssql no_ydb no_vertica no_libsql' \
    github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -p 4 \
    -o ./bin/app \
    cmd/app/main.go

FROM debian:bookworm-slim AS runner

RUN apt-get update && apt-get install -y ca-certificates tzdata make && rm -rf /var/lib/apt/lists/*

ENV TZ=Europe/Minsk
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=builder /usr/local/src/bin/app .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY configs /configs
COPY internal/migrations /migrations

EXPOSE 80

CMD ["./app"]