FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8000 8001 4222 8222 8080

COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

CMD ["/entry.sh"]