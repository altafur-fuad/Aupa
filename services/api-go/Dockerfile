# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies and create non-root user
RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -S aupa && adduser -S aupa -G aupa

# Copy compiled binary from builder
COPY --from=builder /app/api /app/api

USER aupa

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/api"]
