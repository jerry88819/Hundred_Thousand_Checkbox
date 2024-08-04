FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download

RUN go version

RUN go build -o main .

RUN ls -al /app

EXPOSE 8080

CMD ["./main"]
