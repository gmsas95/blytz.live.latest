# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY services/logistics-service/go.mod services/logistics-service/go.sum ./

# Copy service source code
COPY services/logistics-service/ ./

# Copy shared package to correct location for replace directive
COPY ./shared/ ./shared/

# Download dependencies
RUN go mod download

# Update go.sum with shared package dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o logistics-service .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/logistics-service .

# Expose port
EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3   CMD curl -f http://localhost:8087/health || exit 1

# Run the application
CMD ["./logistics-service"]
