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

## 🚀 Production-Ready Scaling (Assignment 4 Updates)

### 1. Cache-Aside & Atomic Invalidation
The Order Service utilizes **Redis** to alleviate database read pressure. 
* **Read Path:** When fetching an order by ID, the system queries Redis first. If a cache miss occurs, it queries PostgreSQL and immediately populates the Redis cache with a 5-minute TTL.
* **Invalidation Strategy:** Cache consistency is strictly maintained. The moment an order's state changes in PostgreSQL (e.g., transitioning from "Pending" to "Paid" via the Payment gateway), the Order Service aggressively issues a `DEL` command to Redis for that specific `orderID`. This guarantees that subsequent reads fetch the fresh state from the database.

### 2. Background Worker & External Adapters
The Notification Service has evolved into a robust Background Worker completely detached from the user's API response time.
* **Adapter Pattern:** External APIs (like SMTP or Mailjet) are hidden behind an `EmailSender` interface. The implementation can be swapped via the `PROVIDER_MODE` environment variable without altering business logic.
* **Retry Logic & Exponential Backoff:** External providers are inherently flaky. If the provider returns a transient error, the worker catches it and retries the job. To avoid overwhelming the recovering external service, it uses an **Exponential Backoff** strategy, sleeping for `2s`, then `4s`, then `8s` before making subsequent attempts.
* **Distributed Idempotency:** To prevent duplicate emails during retries or network blips, the worker uses Redis. It calls `SETNX` (Set if Not Exists) using the event ID. If the key exists, the worker safely ACKs the message and skips processing.

### 3. Bonus: API Rate Limiter
The Order Service is protected by a Redis-backed Rate Limiter middleware. It tracks the incoming IPs, using Redis's `INCR` and `EXPIRE` commands to allow a maximum of 10 requests per minute. Exceeding this limit returns an HTTP `429 Too Many Requests`, protecting the core infrastructure from spikes.

### 1. Bring up the containers
```bash
docker-compose up --build
