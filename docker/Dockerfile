# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY ../go.mod go.sum ./
RUN go mod download

# Copy source code
COPY .. .

# Install golang-migrate CLI
RUN go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Build the application from cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o server ./cmd/server.go

RUN echo "=== Builder Stage File Structure ===" && \
    echo "Working directory: $(pwd)" && \
    ls -la && \
    echo "" && \
    echo "=== db/migrations ===" && \
    ls -la ./db/migrations && \
    echo "" && \
    echo "=== docs ===" && \
    ls -la ./docs || echo "docs not found"

# Run stage
FROM alpine:3.20

# 필수 패키지 설치
RUN apk --no-cache add ca-certificates tzdata wget curl make

# 비root 유저 생성
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# ⭐ 작업 디렉토리 설정 (USER 전에!)
WORKDIR /app

# ⭐ 바이너리 복사
COPY --from=builder --chown=appuser:appgroup /app/server .

# ⭐ golang-migrate CLI 복사
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# ⭐ Makefile 복사 (make migrate-up 등 실행용)
COPY --from=builder --chown=appuser:appgroup /app/Makefile .

# ⭐ db 디렉토리 복사 (migrations 포함)
COPY --from=builder --chown=appuser:appgroup /app/db ./db

# ⭐ docs 디렉토리 복사 (Swagger)
COPY --from=builder --chown=appuser:appgroup /app/docs ./docs

# ⭐ 런타임 파일 구조 확인
RUN echo "=== Runtime File Structure ===" && \
    echo "Working directory: $(pwd)" && \
    ls -laR /app && \
    echo "" && \
    echo "=== Detailed db/migrations ===" && \
    ls -la /app/db/migrations && \
    echo "" && \
    echo "=== Checking permissions ===" && \
    ls -la /app/server && \
    echo "" && \
    echo "=== User info ===" && \
    id appuser

# ⭐ 유저 전환 (마지막!)
USER appuser

# 포트 노출
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 실행
CMD ["./server"]