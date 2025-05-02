FROM golang:latest

RUN mkdir /app
WORKDIR /app
ADD . /app

RUN go install -mod=mod github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main