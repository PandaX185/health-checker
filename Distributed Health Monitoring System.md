# Distributed Health Monitoring System

## Objective
Design and implement a **distributed system** that monitors the health of external services by performing scheduled probes over **REST (HTTP)**.

The system must process checks asynchronously, persist results reliably, and broadcast real-time status updates.

This task evaluates your ability to design **production-ready backend systems**, reason about scalability, concurrency, and reliability, and write clean, testable code.

---

## Requirements

### Functional Requirements

1. **Service Registration**
   - Register external services to monitor:
     - Service name
     - Endpoint URL
     - Check interval (seconds)

2. **Scheduled Health Checks**
   - Periodically execute health checks based on the configured interval.
   - Perform REST-based HTTP checks (e.g. GET `/health`).
   - Each check must have a configurable timeout.

3. **Asynchronous Processing**
   - Health checks must be dispatched via a **Message Queue**.
   - Workers consume jobs and execute probes independently of the scheduler.

4. **Result Storage**
   - Persist health check results in PostgreSQL:
     - status (UP / DOWN)
     - response time
     - timestamp
   - Store results as **append-only logs**.

5. **Real-Time Broadcasting**
   - Broadcast service status changes (UP ↔ DOWN) to connected clients via **WebSockets**.
   - Do not broadcast every check — only **state transitions**.

6. **API Endpoints**
   - Register / list monitored services
   - Query historical health data
   - WebSocket endpoint for live updates

---

### Optional Features (Bonus, Not Required)

- **gRPC Health Checks**
  - Support monitoring gRPC services using a simple health or ping method
- Failure threshold before marking a service as DOWN
- Pause / resume monitoring for a service
- Service tags or groups
- Basic authentication for APIs

---

## Non-Functional Requirements

### Performance
- Support monitoring **thousands of services**
- Scheduler must not block on probe execution
- Workers must use timeouts to avoid hanging probes

### Scalability
- Stateless application design
- Horizontal scaling of workers
- Message queue decouples scheduling from execution

### Reliability
- Failed probes must not crash the system
- At-least-once job processing is acceptable
- Duplicate probe results should be safe to store

### Observability
- Structured logging (Zap / Zerolog / Logrus)
- Include:
  - service name
  - probe result
  - latency
  - error (if any)

---

## Testing Requirements

- **80%+ code coverage**
- Unit tests for:
  - scheduler logic
  - worker processing
  - state transition logic
- Integration tests for:
  - database persistence
  - message queue interaction
- No need to test WebSocket protocol internals

---

## Deployment Requirements

- Containerize the application using **Docker**
- Provide a **docker-compose** setup including:
  - application
  - PostgreSQL
  - message queue (e.g., RabbitMQ or Redis)

---

## Technical Stack

- **Language**: Go (latest stable version)
- **Framework**: Standard library or lightweight (Gin / Chi)
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ / Redis Streams / similar
- **Realtime**: WebSockets
- **Testing**: Go `testing`, `testify`
- **Deployment**: Docker, docker-compose

---

## Deliverables

1. **Git Repository**
   - Clean project structure
   - Meaningful commit history

2. **Documentation**
   - `README.md` or `docs/architecture.md` including:
     - Architecture overview
     - Data flow explanation
     - Key design decisions and trade-offs
     - Local setup instructions

3. **API Documentation**
   - List endpoints and WebSocket events
   - Example request / response payloads

---

## Evaluation Criteria

- **Architecture Quality**
  - Clear separation between scheduler, workers, storage, and broadcasting
- **Code Quality**
  - Idiomatic Go
  - Proper error handling
  - Minimal but meaningful comments
- **Design Decisions**
  - Reasonable trade-offs
  - Avoidance of unnecessary complexity
- **Testing**
  - Focus on business logic, not infrastructure trivia
- **Documentation**
  - Clear explanation of how the system works and why choices were made

---

## Submission Guidelines

- Provide a public or shared Git repository
- Include setup instructions to run locally using `docker-compose`
- Expected completion time: **72 hours**

---

## Notes

- Focus on **clarity and correctness**, not feature quantity
- Prefer simple, reliable solutions over complex abstractions
- The goal is to demonstrate how you think about backend systems under realistic constraints

Good luck — we’re interested in your design choices as much as your code.
