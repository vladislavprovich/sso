FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o sso ./cmd/sso

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/sso .

EXPOSE 8080

CMD ["./sso"]