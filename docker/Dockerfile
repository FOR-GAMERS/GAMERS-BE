# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application from cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server.go

# Run stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

WORKDIR /home/appuser

COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /home/appuser

# Switch to non-root user
USER appuser

EXPOSE 8080

CMD ["./main"]