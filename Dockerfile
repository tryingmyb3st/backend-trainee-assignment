FROM golang:1.26.0-alpine3.22

WORKDIR /task

COPY . .

RUN go mod download

RUN go build -o task ./cmd/task/task.go

EXPOSE 8080

CMD ["./task"]