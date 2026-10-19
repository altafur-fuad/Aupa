# Stage 1: Build Go binary from repository root
FROM golang:1.26-alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates tzdata

# Copy dependency specifications from services/api-go
COPY services/api-go/go.mod services/api-go/go.sum ./
RUN go mod download

# Copy services/api-go source code
COPY services/api-go/ ./

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Stage 2: Lightweight runtime image
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -S aupa && adduser -S aupa -G aupa

COPY --from=builder /app/api /app/api

USER aupa

EXPOSE 8080 10000

ENV PORT=10000
ENV ENVIRONMENT=production

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

ENTRYPOINT ["/app/api"]
