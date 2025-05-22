## 🛠 Prerequisites

Before you begin, make sure the following tools are installed:

* **Go 1.24.2 or later**: [Download Go](https://golang.org/dl/)
* **Docker & Docker Compose**: [Download Docker Desktop](https://www.docker.com/products/docker-desktop)

---

## 🚀 Getting Started

Follow these steps to set up and run the project locally:

1. **Clone the Repository**

   ```bash
   git clone <your-repo-url>
   cd <your-project-directory>
   ```

2. **Copy and Configure Environment Variables**

   ```bash
   cp .env.example .env
   ```

   Fill in the required environment values provided by the admin.

3. **Start Docker Services**

   ```bash
   docker compose up --build -d
   ```

4. **Run Tests**

   Run the tests inside the terminal. Make sure NATS is running in Docker.

   ```bash
   go test ./services/horizon_test
   ```

   > ⏳ *Tests may take 1 to 5 minutes to complete.*

5. **Run the Server**

   ```bash
   go run main.go
   ```

   This will also **auto-migrate the database** and **seed** the initial data.
