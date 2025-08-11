Here's an improved and cleaner version of your `README.md`, with better structure, consistent formatting, and clearer explanations:

---

# 🌀 E-Coop Server

E-Coop Server is the backend system for the E-Coop platform. It’s built with **Go** and optimized to run in a **Dockerized environment** for easy deployment and scalability.

---

## 🛠️ Prerequisites

Before you begin, ensure you have the following installed:

- **Go**: `v1.24.3` or later
- **Docker** and **Docker Compose**

## 🚀 Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/Lands-Horizon-Corp/e-coop-server.git e-coop-server
cd e-coop-server
```

### 2. Configure Environment

Copy the example `.env` file and update values as needed:

```bash
cp .env.example .env
```

---

## 🧑‍💻 Running the Application

### 3. Start Services

Build and start all required services (DB, cache, broadcaster, etc.):

```bash
docker compose up --build -d
```

### 4. Verify Setup

Run tests to ensure the environment is working:

```bash
go clean -cache && go test -v ./services/horizon_test
```

### 5. (Optional) Clean Cache

```bash
go run . cache:clean
```

---

## 🗄️ Database Management

### Automigrate all tables:

```bash
go run . db:migrate
```

### Seed the database:

```bash
go run . db:seed
```

### Reset the database (⚠️ Deletes all data, seeds, and re-migrates):

```bash
go run . db:reset
```

---

## 🧩 Run the Main Server

```bash
go run . server
```

Then visit:

```
http://localhost:8000/routes
```

---

## ❗ Troubleshooting: Port Issues

If you encounter issues with ports already in use:

```bash
chmod +x kill_ports.sh
./kill_ports.sh
```

---

## 🚢 Deployment

### Prepare Code for Deployment

```bash
# Make sure formatting and linting is clean
export PATH="$PATH:$HOME/go/bin"

goimports -w .
gofmt -w .
golangci-lint run
```

### Deploy to Fly.io, Reset Machines & View Logs

```bash
fly deploy; fly machine restart 148e4d55f36278; fly machine restart 90802d3ea0ed38; fly logs
```

https://cooperatives-development.fly.dev/api/v1/transaction/search/branch/search?filter=eyJmaWx0ZXJzIjpbXSwibG9naWMiOiJBTkQifQ%3D%3D&pageIndex=0&pageSize=10&sort=W10%3D