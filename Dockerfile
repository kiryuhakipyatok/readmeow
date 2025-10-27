# syntax=docker/dockerfile:1.6
FROM golang:1.25-alpine AS builder

WORKDIR /usr/local/src

RUN apk add -U --no-cache ca-certificates

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

FROM alpine AS runner

RUN apk add --no-cache tzdata

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV TZ=Europe/Moscow

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apk add --no-cache make

WORKDIR /app

COPY --from=builder /usr/local/src/bin/app .

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY configs /configs

COPY internal/migrations /migrations

EXPOSE 3333

CMD ["./app"]