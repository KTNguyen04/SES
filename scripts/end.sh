#!/usr/bin/env bash

echo "Stopping all ./bin/main processes..."

# Tìm và kill tất cả process ./bin/main
pkill -f "./bin/main"

echo "All processes stopped."
