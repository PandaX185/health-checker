# Health Checker - Distributed Health Monitoring System

A production-ready distributed system for monitoring external service health with real-time WebSocket updates.

## 🚀 Features

- **Service Registration**: Register external services with custom check intervals
- **Asynchronous Health Checks**: Distributed workers process checks via Redis streams
- **Real-time Updates**: WebSocket broadcasting for status changes (UP ↔ DOWN)
- **Historical Data**: PostgreSQL storage with append-only logs
- **REST API**: Full CRUD operations with JWT authentication
- **Swagger Documentation**: Interactive API documentation
- **Docker Support**: Containerized deployment with Docker Compose

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Scheduler     │    │     Redis       │    │     Worker      │
│                 │    │   Streams       │    │                 │
│ • Claims due    │───▶│ • Message Queue │───▶│ • HTTP Checks   │
│   services      │    │ • Job Queue     │    │ • Status Updates │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  PostgreSQL     │    │   WebSocket     │    │     Clients     │
│                 │    │                 │    │                 │
│ • Service Data  │    │ • Real-time     │◀───│ • Browsers      │
│ • Health Logs   │    │   Updates       │    │ • Dashboards    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Tech Stack

- **Backend**: Go 1.24
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Cache/Queue**: Redis
- **Authentication**: JWT
- **WebSockets**: golang.org/x/net/websocket
- **Documentation**: Swagger/OpenAPI
- **Logging**: Zap
- **Testing**: Testify

## 📋 Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL (or use Docker)
- Redis (or use Docker)

## 🚀 Quick Start

### 1. Clone the repository
```bash
git clone <repository-url>
cd health-checker
```

### 2. Start dependencies with Docker
```bash
docker-compose up -d redis
```

### 3. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Run the application
```bash
# Development with hot reload
air

# Or build and run
go build -o bin/health-checker cmd/main.go
./bin/health-checker
```

### 5. Access the application
- **API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html

## ⚙️ Configuration

Create a `.env` file in the project root:

```env
# Environment
ENV=development

# Database
DATABASE_URL=postgres://user:password@localhost:5432/health_checker?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your_super_secret_jwt_key_here

# Server
PORT=:8080
```

## 📖 API Documentation

### Authentication

All API endpoints require JWT authentication except login/register.

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Service Management

```bash
# Register a service to monitor
curl -X POST http://localhost:8080/api/v1/services \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API",
    "url": "https://api.example.com/health",
    "check_interval": 60
  }'

# List all services
curl -X GET http://localhost:8080/api/v1/services \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get health check history
curl -X GET "http://localhost:8080/api/v1/services/1/health-checks?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Real-time WebSocket Updates

Connect to receive live status change notifications:

```bash
# Using wscat (install with: npm install -g wscat)
wscat -c ws://localhost:8080/api/v1/services/ws \
  --origin http://localhost:8080 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Example WebSocket messages:**
```json
{
  "ServiceID": 1,
  "OldStatus": "UP",
  "NewStatus": "DOWN",
  "Timestamp": "2025-12-31T14:30:00Z"
}
```

## 🧪 Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
# Start dependencies
docker-compose up -d

# Run integration tests
go test -tags=integration ./...
```

### WebSocket Testing
```bash
# Install wscat for WebSocket testing
npm install -g wscat

# Connect and monitor status changes
wscat -c ws://localhost:8080/api/v1/services/ws \
  --origin http://localhost:8080 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ Development

### Project Structure
```
health-checker/
├── cmd/                    # Application entrypoints
│   └── main.go
├── internal/
│   ├── app/               # HTTP server and middleware
│   │   ├── auth/          # Authentication handlers
│   │   └── server.go
│   ├── database/          # Database connections
│   ├── logger/            # Logging utilities
│   ├── middleware/        # HTTP middleware
│   ├── migrations/        # Database migrations
│   └── monitor/           # Core monitoring logic
│       ├── domain.go      # Data models
│       ├── handler.go     # HTTP handlers
│       ├── repository.go  # Data access layer
│       ├── service.go     # Business logic
│       ├── scheduler.go   # Job scheduling
│       ├── worker.go      # Background workers
│       ├── hub.go         # WebSocket hub
│       ├── client.go      # WebSocket clients
│       └── event_bus.go   # Event publishing
├── docs/                  # Generated documentation
├── docker-compose.yml     # Docker services
├── go.mod                 # Go modules
└── README.md
```

### Adding New Features

1. **Database Changes**: Add migrations in `internal/migrations/`
2. **API Endpoints**: Add handlers in `internal/monitor/handler.go`
3. **Business Logic**: Add services in `internal/monitor/service.go`
4. **Real-time Events**: Publish to event bus in relevant services

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Generate Swagger docs
swag init -g cmd/main.go
```

## 🚀 Deployment

### Docker Compose (Development)
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f health-checker
```

### Production Deployment

1. **Build the application**
```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go
```

2. **Create production Docker image**
```dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

3. **Deploy with your preferred orchestration tool** (Kubernetes, Docker Swarm, etc.)

## 📊 Monitoring & Observability

- **Logs**: Structured JSON logging with Zap
- **Health Checks**: Built-in service monitoring
- **Metrics**: Ready for integration with Prometheus/Grafana

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Gin Web Framework](https://gin-gonic.com/)
- Database operations with [pgx](https://github.com/jackc/pgx)
- Redis client with [go-redis](https://github.com/go-redis/redis)
- WebSocket support with [golang.org/x/net](https://golang.org/x/net)

---

**Happy monitoring! 🔍**