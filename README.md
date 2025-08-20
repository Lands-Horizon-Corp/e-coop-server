<div align="center">
  <img src="assets/logo.png" alt="E-Coop Server Logo" width="200"/>
</div>

E-Coop Server is a server for multipurpose cooperatives. A comprehensive financial cooperative management system built with Go. The backend API server provides robust account management, transaction processing, and organizational tools for cooperative financial institutions.

## Prerequisites

- **Go** 1.25.0 or later
- **Docker** and **Docker Compose**
- **PostgreSQL** 13+
- **Redis** 6+

## Quick Start

### Environment Setup

```bash
git clone https://github.com/Lands-Horizon-Corp/e-coop-server.git
cd e-coop-server
cp .env.example .env
```

### Start Infrastructure

```bash
docker compose up --build -d
```

### Database Operations

```bash
# Migrate database schema
go run . db migrate

# Seed initial data
go run . db seed

# Reset and refresh database
go run . db refresh

# Reset database (drops all tables)
go run . db reset
```

### Cache Management

```bash
# Clean application cache
go run . cache clean
```

### Start Server

```bash
go run . server
```

The server will be available at `http://localhost:8000`

## API Documentation

Visit `http://localhost:8000/routes` for interactive API documentation.

## Commands Reference

### Database Commands

| Command               | Description                             |
| --------------------- | --------------------------------------- |
| `go run . db migrate` | Auto-migrate all database tables        |
| `go run . db seed`    | Populate database with initial data     |
| `go run . db reset`   | Drop and recreate all tables            |
| `go run . db refresh` | Reset database and seed with fresh data |

### Cache Commands

| Command                | Description           |
| ---------------------- | --------------------- |
| `go run . cache clean` | Flush all cached data |

### Server Commands

| Command           | Description                       |
| ----------------- | --------------------------------- |
| `go run . server` | Start the main application server |

### Utility Commands

| Command            | Description                 |
| ------------------ | --------------------------- |
| `go run . version` | Display version information |
| `go run . --help`  | Show available commands     |

## Development

### Code Quality

```bash
# Format code
goimports -w .
gofmt -w .

# Run linter
golangci-lint run

# Run tests
go test -v ./services/horizon_test
```

### Port Management

If encountering port conflicts:

```bash
chmod +x kill_ports.sh
./kill_ports.sh
```

## Deployment

### Pre-deployment Checks

```bash
export PATH="$PATH:$HOME/go/bin"
goimports -w .
gofmt -w .
golangci-lint run
```

### Deploy to Fly.io

```bash
fly deploy
fly machine restart 148e4d55f36278
fly machine restart 90802d3ea0ed38
fly logs
```

## Architecture

The server follows a clean architecture pattern with:

- **Controllers** - HTTP request handling
- **Services** - Business logic layer
- **Models** - Data access layer
- **Middleware** - Cross-cutting concerns

## License

This project is proprietary software owned by Lands Horizon Corp.
