# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod
COPY services/payment-service/go.mod ./

# Copy go.sum if it exists
COPY services/payment-service/go.sum* ./

# Copy shared package
COPY shared/ ./shared/

# Copy service source code
COPY services/payment-service/ ./

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment-service .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S -u 1001 -G appuser appuser

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/payment-service .

# Change ownership to non-root user
RUN chown appuser:appuser payment-service

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8089

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8089/health || exit 1

# Run the application
CMD ["./payment-service"]