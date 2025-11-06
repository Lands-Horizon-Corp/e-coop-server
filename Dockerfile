FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Only copy go.mod and go.sum first for better caching of dependencies
COPY go.mod go.sum ./
RUN go mod download

RUN go clean -modcache


# Now copy the rest of the source code
COPY . .

# Build the Go binary statically and strip debug info
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app .

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