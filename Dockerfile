# =============================================================================
# STAGE 1: Build
# =============================================================================
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git=2.52.0-r0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o server ./cmd/

# =============================================================================
# STAGE 2: Runtime
# =============================================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates=20250911-r0

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 3000

# No entrypoint.sh needed — nothing to do before the app starts
CMD ["/app/server"]
