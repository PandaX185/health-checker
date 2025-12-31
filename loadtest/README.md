# Load Testing Tool

This tool is designed to verify that the Health Monitoring System can handle a large number of services.

## Prerequisites

1. The Health Monitoring System must be running (e.g., via `docker-compose up` or locally).
2. Go must be installed to run the load test.

## Usage

Run the load test from the project root:

```bash
go run loadtest/main.go [flags]
```

### Available Flags

- `-api`: Base URL of the Health Checker API (default: "http://localhost:8080/api/v1")
- `-services`: Number of services to register (default: 1000)
- `-concurrency`: Number of concurrent workers (default: 20)
- `-mock-port`: Port for the mock service to listen on (default: 9090)

### Example

Register 5000 services with 50 concurrent workers:

```bash
go run loadtest/main.go -services 5000 -concurrency 50
```

## What it does

1. **Starts a Mock Server**: Listens on the specified port (default 9090) to act as the target for health checks.
2. **Registers Services**: Sends POST requests to the API to register the specified number of services. Each service is configured to check `http://localhost:<mock-port>/health/<id>`.
3. **Measures Performance**: Reports the time taken, throughput, and success/failure counts.
4. **Waits for Checks**: Keeps running for 30 seconds to allow the system to perform health checks against the mock server (you should see logs in the main application).

## Notes

- If running the application in Docker, ensure it can reach the host machine if you use `localhost` in the service URLs. You might need to use `host.docker.internal` or the host's IP address if the load test is running on the host and the app is in a container.
- To test with Docker, you might want to run this load test tool inside the docker network or configure the service URLs appropriately.
