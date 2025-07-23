# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git postgresql-client bash

ENV GOBIN=/app/bin
ENV PATH=$GOBIN:$PATH

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN echo "check goose" && ls -l /app/bin

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o usdt-service ./cmd/app/main.go


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache postgresql-client bash ca-certificates


COPY --from=builder /app/bin/goose /usr/local/bin/goose
COPY --from=builder /app/usdt-service .

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 50051 2112

HEALTHCHECK --interval=30s --timeout=3s \
  CMD nc -z localhost 2112 || exit 1

ENTRYPOINT ["./entrypoint.sh"]
