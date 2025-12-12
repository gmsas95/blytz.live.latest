FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY services/notification-service/go.mod services/notification-service/go.sum ./

# Copy shared package
COPY shared/ ./shared/

# Copy service source code
COPY services/notification-service/ ./

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o notification-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/notification-service .

# Expose port
EXPOSE 8094

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8094/health || exit 1

# Run the binary
CMD ["./notification-service"]