# 📦 E-Commerce Platform: Event-Driven Microservices

This platform is a resilient, distributed system built with **Go**, **gRPC**, **PostgreSQL**, and **RabbitMQ**. It demonstrates the evolution from tightly coupled synchronous services to an **Event-Driven Architecture**, ensuring high maintainability, fault tolerance, and message reliability.

---

## 🏛 Architecture Decisions

### 1. Clean Architecture
Each microservice (Order, Payment, Notification) is structured using Clean Architecture principles, partitioned into distinct layers: **Domain, Use Case, Repository, and Transport/Broker**.
* **Composition Root:** Dependency injection is handled manually in `main.go` to keep business logic independent of external frameworks.

### 2. Event-Driven Flow & Decoupling
To prevent synchronous bottlenecks and third-party API dependencies, the platform uses an asynchronous flow:
* **Payment Service (Producer):** After a successful database transaction, it publishes a `payment.completed` event to RabbitMQ.
* **Notification Service (Consumer):** A completely decoupled service that listens to the RabbitMQ queue and simulates sending emails without any direct knowledge of the Order or Payment services.

---

## ⚡ Core Features & Reliability Guarantees

### 1. Delivery Guarantees & ACK Logic
We achieve **At-Least-Once delivery** by ensuring messages are never lost if a consumer crashes mid-processing:
* **Manual ACKs:** The `auto-ack` setting is completely disabled in the RabbitMQ consumer.
* **Confirmation Check:** A message is only explicitly acknowledged (`msg.Ack(false)`) after the simulated email log successfully prints to the console and the state is saved. If the Notification Service crashes before this, the broker detects the dropped TCP connection and requeues the message.

### 2. Idempotency Strategy
Because At-Least-Once delivery can result in duplicate messages (e.g., network retries), the consumer must be idempotent:
* **Mechanism:** Every incoming event contains a unique `EventID`. The Notification Service utilizes an in-memory cache (`sync.Map`) as its data store to track processed IDs.
* **Logic:** Before processing a message, the service checks the cache. If the ID exists, the message is immediately acknowledged and safely discarded. If it doesn't exist, the email is simulated, the ID is recorded, and the message is explicitly acknowledged.

### 3. Advanced Failure Handling: Dead Letter Queue (DLQ)
To handle poison messages or permanent failures without clogging the main queue, the broker is configured with a Dead Letter Exchange (DLX):
* If the consumer encounters an unrecoverable error (simulated via `OrderID == "FAIL_ME"`), the message is explicitly rejected (`msg.Nack(false, false)`).
* RabbitMQ automatically routes this rejected message to the dedicated `payment_dlq` for later inspection.

---

## 🚀 Setup & Testing

### 1. Bring up the containers
```bash
docker-compose up --build