# Use Debian for builder to avoid OOM
FROM golang:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GOMAXPROCS=1
ENV GOPARALLEL=1

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o app .

# Runtime image (Alpine)
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ffmpeg

COPY --from=builder /app/app .
COPY --from=builder /app/seeder ./seeder
COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

EXPOSE 8000 8001 4222 8222 8080

CMD ["/entry.sh"]
