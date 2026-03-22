# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install CA certs and git (needed for go mod download)
RUN apk --no-cache add ca-certificates git

# Download dependencies (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -ldflags="-w -s" -o server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# Non-root user
RUN adduser -D -H -h /app byteroom
USER byteroom

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./server"]
