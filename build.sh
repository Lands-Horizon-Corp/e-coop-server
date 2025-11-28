#!/bin/bash

# Build script for local testing before Docker deployment
set -e

echo "🔧 Starting local build test..."

# Set build constraints
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOMEMLIMIT=1GiB

echo "📦 Downloading dependencies..."
go mod download

echo "🧹 Cleaning build cache..."
go clean -cache -modcache -testcache -fuzzcache

echo "🏗️  Building application..."
go build -a -installsuffix cgo -ldflags="-s -w" -o app .

echo "✅ Build completed successfully!"

# Check binary size
if [ -f "app" ]; then
    SIZE=$(ls -lh app | awk '{print $5}')
    echo "📊 Binary size: $SIZE"
fi

echo "🐳 Ready for Docker build!"