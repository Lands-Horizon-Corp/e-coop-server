FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY . .
COPY .env .

RUN go mod download

RUN go build -o app .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/.env .env         
COPY ./config/nats.conf ./config/nats.conf
EXPOSE 8000 8001 4222 8222 8080

RUN apk add --no-cache curl tar && \
    curl -L https://github.com/nats-io/nats-server/releases/download/v2.11.4/nats-server-v2.11.4-linux-amd64.tar.gz | tar xz && \
    mv nats-server-v2.11.4-linux-amd64/nats-server /usr/local/bin/ && \
    rm -rf nats-server-v2.11.4-linux-amd64

COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

CMD ["/entry.sh"]