FROM golang:1.26.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o torii ./cmd/torii

FROM debian:stable-slim AS runtime-base

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

VOLUME /data
WORKDIR /data
EXPOSE 27000

ENTRYPOINT ["/torii"]
CMD ["--config", "/etc/torii/config.yaml", "--data-dir", "/data"]


FROM runtime-base AS source-build

COPY --from=builder /app/torii /torii

FROM runtime-base AS pre-built

COPY torii /torii