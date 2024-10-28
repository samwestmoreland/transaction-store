FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o transaction-store .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/transaction-store /app/config.yaml ./

CMD ["./transaction-store"]
