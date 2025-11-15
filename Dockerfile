FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o delayed-notifier cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/delayed-notifier .
COPY configs/config.yaml configs/
COPY .env .

EXPOSE 8080

CMD ["./delayed-notifier"]