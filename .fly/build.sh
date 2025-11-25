#!/bin/bash

# Fly.io custom build script
set -e

echo "🚀 Starting Fly.io optimized build..."

# Set environment variables for memory-constrained build
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOMEMLIMIT=1500MiB
export GOGC=50

# Download dependencies first
echo "📦 Downloading Go modules..."
go mod download
go mod verify

# Clean up to free memory
echo "🧹 Cleaning build cache..."
go clean -cache -modcache -testcache -fuzzcache

# Build with memory optimizations
echo "🏗️  Building application..."
go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -X main.version=$(git rev-parse --short HEAD)" \
    -trimpath \
    -o app \
    .

echo "✅ Build completed successfully!"

# Verify binary
if [ -f "app" ]; then
    echo "📊 Binary info:"
    ls -lh app
    file app
fi