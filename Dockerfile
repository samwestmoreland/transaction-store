FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o transaction-store .

FROM alpine:latest

RUN addgroup -g 10000 appgroup && \
    adduser -u 10000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

COPY --from=builder /app/transaction-store /app/config.yaml ./

RUN chown -R appuser:appgroup /app

CMD ["./transaction-store"]
