FROM golang:1.25.0-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd

FROM alpine:latest
WORKDIR /
COPY --from=builder /src/server ./server
COPY ./migrations ./migrations
COPY ./.env .env

CMD ["./server"]