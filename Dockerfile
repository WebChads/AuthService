FROM golang:1.23.6 AS builder

WORKDIR /build
COPY . .
RUN go mod tidy
RUN go build -o auth_service

CMD ["./auth_service"]
