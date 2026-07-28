FROM golang:1.25.12-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Limit parallelism so Render/small builders don't OOM during compile.
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=1 go build -p 1 -o /api ./cmd/api

FROM gcr.io/distroless/static-debian12

WORKDIR /
COPY --from=builder /api /api
COPY --from=builder /app/migrations /migrations

ENV HTTP_PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/api"]
