# E-Coop Server

E-Coop Server is the backend server for the E-Coop platform, built with Go and designed to run in a Dockerized environment.

## 🛠 Prerequisites

Before you begin, ensure you have the following installed on your machine:

- **Go**: Version 1.24.3 or later
- **Docker** and **Docker Compose**

---

## 🚀 Installation and Setup

Follow the steps below to set up and run the E-Coop Server:

### 1. Clone the Repository

```bash
git clone https://github.com/Lands-Horizon-Corp/e-coop-server.git e-coop-server
cd e-coop-server
```

### 2. Configure the Environment

Copy the example environment file and configure it as needed:

```bash
cp .env.example .env
```

---

## 🧑‍💻 Running the Application

### 3. Start Server Services

Run the following command to start the required services (like database, cache, and broadcaster):

```bash
docker compose up --build -d
```

### 4. Test Environment Setup

Verify the environment is working with the running Docker services:

```bash
go clean -cache && go test -v ./services/horizon_test
```

### 5. Clean Cache (Optional)

Clean the cache if needed:

```bash
go run . cache:clean
```

### 6. Database Management

#### Automigrate All Tables:
```bash
go run . db:migrate
```

#### Seed the Database:
```bash
go run . db:seed
```

#### Reset the Database (Optional):
```bash
go run . db:reset
```

### 7. Run the Main Server

Start the server:

```bash
go run main.go
```

### 8. Visit & view all available routes
```
http://localhost:8000/routes
```