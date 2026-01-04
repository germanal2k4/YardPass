# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Build Bot binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# API runtime stage
FROM alpine:latest AS api

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./api"]

# Bot runtime stage
FROM alpine:latest AS bot

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bot .

CMD ["./bot"]

