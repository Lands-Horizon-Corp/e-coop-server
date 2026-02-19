

<div align="center">
  <img src="assets/logo.png" alt="E-Coop Server Logo" width="200"/>
</div>

# E-Coop Server

E-Coop Server is a multipurpose cooperative management system built with **Go**. It provides account management, transaction processing, and organizational tools for financial cooperatives.

---

## Prerequisites

* **Go** 1.26.0+
* **Docker** & **Docker Compose**
* **PostgreSQL** 13+
* **Redis** 6+

---

## Quick Start

### Clone Repository & Setup Environment

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
go run . db-migrate     # Apply database migrations
go run . db-seed        # Seed initial data
go run . db-refresh     # Reset and seed database
go run . db-reset       # Drop and recreate all tables
```

### Cache Management

```bash
go run . cache clean    # Clear application cache
```

### Start Server

```bash
go run . server
```

Server available at: `http://localhost:8000`

API docs: `http://localhost:8000/routes`

---

## Commands Reference

### Database Commands

| Command               | Description                             |
| --------------------- | --------------------------------------- |
| `go run . db-migrate` | Auto-migrate all database tables        |
| `go run . db-seed`    | Populate database with initial data     |
| `go run . db-reset`   | Drop and recreate all tables            |
| `go run . db-refresh` | Reset database and seed with fresh data |

### Cache Commands

| Command                | Description           |
| ---------------------- | --------------------- |
| `go run . cache-clean` | Flush all cached data |

### Server Commands

| Command           | Description                       |
| ----------------- | --------------------------------- |
| `go run . server` | Start the main application server |

### Utility Commands

| Command            | Description              |
| ------------------ | ------------------------ |
| `go run . version` | Show version information |
| `go run . --help`  | Show available commands  |

---

## Development

### Code Quality

```bash
goimports -w .
golangci-lint run
go test -v ./services/horizon_test
```

### Nil Pointer Checker

```bash
export PATH="$PATH:$HOME/go/bin"
nilaway ./...
```

### Port Management

```bash
chmod +x kill_ports.sh
./kill_ports.sh
```

---

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
git add . && git commit -m "deploy" && git push
fly deploy
fly machine restart <machine-id>
fly logs
```
---

## Architecture

* **Controllers** - HTTP request handling
* **Services** - Business logic layer
* **Models** - Data access layer
* **Middleware** - Cross-cutting concerns

---

## Makefile Commands

### Quick Start

| Command      | Description                   |
| ------------ | ----------------------------- |
| `make setup` | Setup development environment |
| `make dev`   | Start development server      |
| `make reset` | Reset db-and restart server   |

### Database

| Command           | Description               |
| ----------------- | ------------------------- |
| `make db-migrate` | Apply database migrations |
| `make db-seed`    | Seed initial data         |
| `make db-reset`   | Reset database            |
| `make db-refresh` | Reset and seed database   |
| `make db-setup`   | Migrate + seed database   |

### Docker

| Command               | Description                        |
| --------------------- | ---------------------------------- |
| `make docker-up`      | Start services with Docker Compose |
| `make docker-down`    | Stop all Docker services           |
| `make docker-restart` | Restart all Docker services        |
| `make docker-logs`    | Show Docker container logs         |

### Code Quality

| Command         | Description                        |
| --------------- | ---------------------------------- |
| `make format`   | Format code                        |
| `make lint`     | Run golangci-lint                  |
| `make quality`  | Run all code quality checks        |
| `make test`     | Run tests                          |
| `make coverage` | Generate HTML test coverage report |

### Build

| Command           | Description                     |
| ----------------- | ------------------------------- |
| `make build`      | Build application binary        |
| `make build-prod` | Build production binary         |
| `make clean`      | Clean build artifacts and cache |
| `make clean-all`  | Clean build, Docker, and cache  |

### Deployment

| Command             | Description           |
| ------------------- | --------------------- |
| `make deploy-check` | Pre-deployment checks |
| `make deploy-fly`   | Deploy to Fly.io      |
| `make deploy-logs`  | Show deployment logs  |

### Development Tools

| Command            | Description                         |
| ------------------ | ----------------------------------- |
| `make cache-clean` | Clean application cache             |
| `make kill-ports`  | Kill processes on conflicting ports |
| `make deps`        | Download and tidy dependencies      |
| `make deps-update` | Update dependencies                 |
| `make version`     | Show version information            |
| `make routes`      | Show API routes                     |

### Advanced

| Command          | Description                  |
| ---------------- | ---------------------------- |
| `make benchmark` | Run performance benchmarks   |
| `make mod-graph` | Show module dependency graph |
| `make install`   | Install binary to system     |
| `make uninstall` | Remove binary from system    |

---

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

---

## License

Proprietary software owned by **Lands Horizon Corp.**
