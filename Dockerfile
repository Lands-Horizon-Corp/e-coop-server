FROM golang:latest

# Set up app directory
RUN mkdir /app
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the app
COPY . .

# Install CompileDaemon and swag
RUN go install -mod=mod github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Run swag init to generate docs
RUN swag init

# Run with CompileDaemon
ENTRYPOINT ["CompileDaemon", "--build=go build main.go", "--command=./main"]
