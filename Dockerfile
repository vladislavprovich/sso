FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && go build -o sso-main ./cmd/sso

FROM ubuntu:22.04

WORKDIR /root/

COPY --from=builder app/cmd/sso .

CMD ["./sso-main"]