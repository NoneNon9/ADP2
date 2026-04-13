# 📦 E-Commerce Platform: gRPC Microservices & Clean Architecture

This platform is a resilient, distributed system built with **Go**, **gRPC**, and **PostgreSQL**. It demonstrates **Clean Architecture** within two distinct **Bounded Contexts**, ensuring high maintainability, strict fault tolerance, and real-time streaming capabilities.

---

## 🏛 Architecture Decisions

### 1. Clean Architecture (Internal)
Each service is partitioned into four distinct layers: **Domain, Use Case, Repository, and Transport**.
* **Decision:** We use a **Composition Root** in `main.go` for manual Dependency Injection.
* **Reasoning:** This avoids "magic" frameworks and keeps the business logic (Use Cases) completely independent of the database (Postgres) and the delivery mechanisms (HTTP/REST and gRPC).

### 2. Contract-First Flow (Protobufs)
* **Decision:** All inter-service contracts (`.proto` files) are strictly managed in a separate remote repository: [github.com/NoneNon9/convertedProto](https://github.com/NoneNon9/convertedProto).
* **Reasoning:** This ensures both the Order and Payment services rely on a single source of truth for their communication schemas, fetched automatically via Go Modules (`go.mod`).

### 3. Bounded Contexts & Data Ownership
* **Decision:** Each service has its own dedicated database (`orderdb` and `paymentdb`).
* **Reasoning:** We strictly follow the **Database-per-Service** pattern. This prevents "Hidden Coupling" where one service accidentally depends on the table structure of another.

---

## ⚡ Core Features

### 1. gRPC Inter-Service Communication
* The external API facing the client remains **REST/JSON** (port 8080) for wide accessibility.
* Internal communication between the Order Gateway and Payment Service uses **gRPC** (port 50051) for high-performance, strongly typed binary serialization.

### 2. Real-Time Order Tracking (Server-Side Streaming)
* The Order Service implements a gRPC Server-Side Stream (`SubscribeToOrderUpdates` on port 50052).
* Instead of clients polling for updates, the `OrderUseCase` uses internal Go channels to actively push database state changes directly to connected clients in real-time.

### 3. Idempotency & Failure Handling
* **Idempotency-Key:** The `POST /orders` endpoint supports an idempotency header. If the same key is sent twice (e.g., due to a network retry), the service returns the existing order instead of double-charging.
* **Error Propagation:** Database errors and validation failures are properly mapped to standard `google.golang.org/grpc/status` codes (e.g., `InvalidArgument`, `Internal`).

---

## 🚀 Setup & Testing

### 1. Bring up the containers
```bash
docker-compose up --build