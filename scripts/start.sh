#!/usr/bin/env bash

set -e

CONFIG_FILE="./config.yaml"

# --- Build binary ---
echo "Building Go binary..."
go build -o ./bin/main ./main.go
echo "Build done."

echo "Starting processes..."


PORTS=$(grep 'port:' "$CONFIG_FILE" | awk '{print $2}')

for PORT in $PORTS; do
    echo "Starting process on port $PORT"
    ./bin/main -port="$PORT" &
done

echo "All processes started."
