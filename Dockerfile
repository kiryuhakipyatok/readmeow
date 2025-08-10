FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src

COPY go.mod go.sum ./
RUN go mod download 

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o ./bin/app cmd/app/main.go

FROM alpine AS runner

RUN apk add --no-cache tzdata

ENV TZ=Europe/Moscow

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apk add --no-cache make

WORKDIR /app

COPY --from=builder /usr/local/src/bin/app .

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY .env .

COPY config/config-local.yaml /config/config-local.yaml

COPY internal/migrations /migrations

EXPOSE 3333

CMD ["./app"]