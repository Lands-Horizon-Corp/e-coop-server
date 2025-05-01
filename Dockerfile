FROM golang:latest

RUN mkdir /app
WORKDIR /app
ADD . /app

RUN go get github.com/githubnemo/CompileDaemon
RUN go get github.com/gin-gonic/gin

ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main