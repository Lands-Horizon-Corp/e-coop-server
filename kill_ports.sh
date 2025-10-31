#!/bin/bash

# List of ports to kill
PORTS=(
  8000 8001 4222 8222 8080
  5432 6379 9000 1025 8002
  5540 9001 8025
)

for PORT in "${PORTS[@]}"; do
  echo "🔍 Killing process on port $PORT..."
  sudo fuser -k "${PORT}/tcp"
done

echo "✅ Done.