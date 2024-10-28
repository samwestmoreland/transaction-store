FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o transaction-store .

CMD ["./transaction-store"]
