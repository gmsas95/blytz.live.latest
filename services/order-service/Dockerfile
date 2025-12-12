# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY services/order-service/go.mod services/order-service/go.sum ./

# Copy shared package
COPY ./shared/ ./shared/

# Copy service source code
COPY services/order-service/ ./

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o order-service .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/order-service .

# Expose port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3   CMD curl -f http://localhost:8085/health || exit 1

# Run the application
CMD ["./order-service"]
