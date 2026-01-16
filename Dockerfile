# ---------- Builder ----------
FROM golang:1.25.6-alpine AS builder

# Install git (needed for some modules) and bash
RUN apk add --no-cache git bash

WORKDIR /app

# Copy only go.mod/go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Disable CGO if possible (reduces memory and simplifies build)
ENV CGO_ENABLED=0
ENV GOMAXPROCS=1

# Build binary with reduced size and path trimming
RUN go build -trimpath -ldflags="-s -w" -o app .

# ---------- Runtime ----------
FROM alpine:latest

WORKDIR /app

# Install ffmpeg if required at runtime
RUN apk add --no-cache ffmpeg bash

# Copy built binary and seeder
COPY --from=builder /app/app .
COPY --from=builder /app/seeder ./seeder

# Copy entrypoint script
COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

EXPOSE 8000 8001 4222 8222 8080

CMD ["/entry.sh"]