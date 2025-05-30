FROM golang:1.24.2-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8000
EXPOSE 8001

# Start Air
CMD ["air"]