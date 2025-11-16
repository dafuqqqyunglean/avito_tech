FROM golang:1.25.0-alpine

RUN go version

ENV GOPATH=/

COPY . .

RUN go mod download

RUN go build -o app ./cmd

CMD ["./app"]