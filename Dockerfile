# ── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod first; go.sum is generated inside the build if missing
COPY go.mod ./

# Download and tidy — this regenerates go.sum inside the build container
RUN go mod download

COPY . .

# Tidy again with full source present to catch any missing indirect deps
RUN go mod tidy

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/main.go

# ── Stage 2: minimal runtime image ──────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .
COPY db/migrations ./db/migrations

EXPOSE 8080
CMD ["./server"]
