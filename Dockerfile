FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Only copy go.mod and go.sum first for better caching of dependencies
COPY go.mod go.sum ./
RUN go mod download

# Clean cache to reduce memory footprint
RUN go clean -cache -modcache -testcache -fuzzcache

# Now copy the rest of the source code
COPY . .

# Build the Go binary with memory optimizations
# - Use -gcflags="-N -l" to disable optimizations and reduce memory during compilation
# - Use -ldflags="-s -w" to strip debug info from final binary
# - Set GOMEMLIMIT to control memory usage during build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -ldflags="-s -w -extldflags '-static'" -o app .

# Use a minimal base image for the final artifact
FROM alpine:latest
WORKDIR /app

# Install FFmpeg for image processing
RUN apk add --no-cache ffmpeg

# Copy only the built binary from the builder stage
COPY --from=builder /app/app .

# Copy seeder directory with images for runtime seeding`
COPY --from=builder /app/seeder ./seeder

# Copy entrypoint script and set permissions
COPY entry.sh /entry.sh
RUN chmod +x /entry.sh

# Set Go to use all available CPUs at runtime
ENV GOMAXPROCS=$(nproc)

# Expose only necessary ports
EXPOSE 8000 8001 4222 8222 8080

# Use a non-root user for security (optional, if your app allows)
# RUN adduser -D appuser
# USER appuser

CMD ["/entry.sh"]