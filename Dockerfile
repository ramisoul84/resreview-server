# ── Build stage ───────────────────────────────────────────
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service ./cmd/main.go

# ── Final stage ──────────────────────────────
FROM alpine:3.22 

RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /service /resreview-server

ENTRYPOINT ["/resreview-server"]