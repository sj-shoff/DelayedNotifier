FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o delayed-notifier ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/delayed-notifier .
COPY --from=builder /app/static ./static 
COPY --from=builder /app/.env.docker ./.env

EXPOSE 8080

CMD ["./delayed-notifier"]