# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY services/auction-service/go.mod services/auction-service/go.sum ./

# Copy shared package
COPY ./shared/ ./shared/

# Copy service source code
COPY services/auction-service/ ./

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auction-service .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/auction-service .

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8083/health || exit 1

# Run the application
CMD ["./auction-service"]