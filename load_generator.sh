#!/bin/bash

while true; do
    # Generate a new UUID
    transaction_id=$(uuidgen)

    # Generate a random amount between 1.00 and 1000.00
    amount=$(printf "%.2f" "$(echo "scale=2; $RANDOM%1000 + 0.99" | bc)")

    # Send the POST request
    curl -X POST localhost:8080/api/transaction/ \
        -H "Content-Type: application/json" \
        -d "{\"transactionId\":\"$transaction_id\",\"amount\":\"$amount\",\"timestamp\":\"2009-09-28T19:03:12Z\"}"

    sleep 2
done

