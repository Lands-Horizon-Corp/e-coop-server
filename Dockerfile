# ---------- Builder ----------
FROM golang:1.25.5-bookworm AS builder

# Install native deps for CGO
RUN apt-get update && apt-get install -y \
    build-essential \
    libwebp-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Limit Go build to single core to reduce RAM usage
ENV CGO_ENABLED=1
ENV GOMAXPROCS=1

RUN go build -ldflags="-s -w" -o app .

# ---------- Runtime ----------
FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ffmpeg

COPY --from=builder /app/app .
COPY --from=builder /app/seeder ./seeder
COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

EXPOSE 8000 8001 4222 8222 8080
CMD ["/entry.sh"]
