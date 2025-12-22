<div align="center">
  <img src="assets/logo.png" alt="E-Coop Server Logo" width="200"/>
</div>

E-Coop Server is a server for multipurpose cooperatives. A comprehensive financial cooperative management system built with Go. The backend API server provides robust account management, transaction processing, and organizational tools for cooperative financial institutions.

## Prerequisites

- **Go** 1.25.5 or later
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




# Run linter
golangci-lint run

# Run tests
go test -v ./services/horizon_test
```

### Nil Pointer Checker

```bash
export PATH="$PATH:$HOME/go/bin"

nilaway ./...
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
  git add .; git commit -m "fix: code style, linter, and modernization"; git push; golangci-lint run
  fly deploy; fly machine restart 148e4d55f36278; fly machine restart 90802d3ea0ed38; fly logs
```

## Architecture

The server follows a clean architecture pattern with:

- **Controllers** - HTTP request handlingsd
- **Services** - Business logic layer
- **Models** - Data access layer
- **Middleware** - Cross-cutting concerns

## Makefile Commands

For streamlined development, use the provided Mak dfile commands:

### Quick Start Commands

| Command      | Description                            |
| ------------ | -------------------------------------- |
| `make help`  | Show all available commands            |
| `make setup` | Complete development environment setup |
| `make start` | Quick start (Docker + DB + Server)     |
| `make dev`   | Start development server               |
| `make reset` | Complete reset and restart             |

### Database Commands

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `make db-migrate` | Migrate database schema                  |
| `make db-seed`    | Seed database with initial data          |
| `make db-reset`   | Reset database (drops all tables)        |
| `make db-refresh` | Reset database and seed with fresh data  |
| `make db-setup`   | Complete database setup (migrate + seed) |

### Docker Management

| Command               | Description                            |
| --------------------- | -------------------------------------- |
| `make docker-up`      | Start all services with Docker Compose |
| `make docker-down`    | Stop all Docker services               |
| `make docker-restart` | Restart all Docker services            |
| `make docker-logs`    | Show Docker container logs             |

### Code Quality

| Command           | Description                          |
| ----------------- | ------------------------------------ |
| `make format`     | Format code with goimports and gofmt |
| `make lint`       | Run golangci-lint                    |
| `make quality`    | Run all code quality checks          |
| `make test`       | Run all tests                        |
| `make test-clean` | Run tests with clean cache           |
| `make coverage`   | Generate HTML test coverage report   |

### Build Commands

| Command           | Description                                       |
| ----------------- | ------------------------------------------------- |
| `make build`      | Build the application binary                      |
| `make build-prod` | Build production binary                           |
| `make clean`      | Clean build artifacts and caches                  |
| `make clean-all`  | Clean everything (build artifacts, Docker, cache) |

### Deployment Commands

| Command             | Description                             |
| ------------------- | --------------------------------------- |
| `make deploy-check` | Pre-deployment checks (quality + tests) |
| `make deploy-fly`   | Deploy to Fly.io                        |
| `make deploy-logs`  | Show deployment logs                    |

### Development Tools

| Command            | Description                            |
| ------------------ | -------------------------------------- |
| `make cache-clean` | Clean application cache                |
| `make kill-ports`  | Kill processes using conflicting ports |
| `make deps`        | Download and tidy dependencies         |
| `make deps-update` | Update all dependencies                |
| `make version`     | Show version information               |
| `make routes`      | Show API routes information            |

### Advanced Commands

| Command          | Description                  |
| ---------------- | ---------------------------- |
| `make benchmark` | Run performance benchmarks   |
| `make mod-graph` | Show module dependency graph |
| `make install`   | Install binary to system     |
| `make uninstall` | Remove binary from system    |

### Example Workflows

```bash
# New developer setup
make setup

# Daily development
make dev

# Before committing
make quality
make test

# Production deployment
make deploy-fly
```

## License

This project is proprietary software owned by Lands Horizon Corp.


