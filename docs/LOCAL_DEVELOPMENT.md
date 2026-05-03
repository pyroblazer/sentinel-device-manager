# Local Development Guide

Get the Sentinel Device Manager running locally with a single command.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Docker Compose)

That's it. No Go, Python, Node.js, or AWS CLI needed — everything runs inside containers.

## Quick Start

```bash
make dev
```

This single command:
1. Tears down any previous Docker state (`docker compose down -v`)
2. Builds all service images from scratch (`--build`)
3. Starts DynamoDB Local and creates the `sentinel-devices` table automatically
4. Starts the Go backend, Python analytics, React frontend, and firmware simulator

Alternatively, you can run the Docker commands directly:

```bash
docker compose up --build
```

## Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| Frontend Dashboard | http://localhost:3000 | React dark-mode UI |
| Go REST API | http://localhost:8080 | Device management API |
| Go gRPC | localhost:9090 | gRPC service for devices |
| Python Analytics | http://localhost:8081 | Event and alert service |
| DynamoDB Shell | http://localhost:8000/shell | DynamoDB Local web UI |
| Swagger/OpenAPI | http://localhost:8080/swagger.json | API specification |

## Verifying Everything Works

```bash
# Go service health
curl http://localhost:8080/health
# Expected: {"status":"healthy","version":"1.0.0"}

# Python service health
curl http://localhost:8081/health
# Expected: {"status":"ok"}

# Frontend
curl -s http://localhost:3000 | head -5
```

The firmware simulator automatically registers a virtual camera (`VKD-CAM-SIM-001`) and starts sending heartbeats and sensor events.

## Stopping All Services

```bash
make dev-down
```

Or directly:

```bash
docker compose down -v
```

The `-v` flag removes DynamoDB data volumes so the next start is completely fresh.

## Running with Observability

To include Prometheus, Grafana, Elasticsearch, Kibana, and the OTel Collector:

```bash
make obs-up
```

This starts all application services plus the full observability stack:

| Service | URL |
|---------|-----|
| Prometheus | http://localhost:9091 |
| Grafana | http://localhost:3001 (admin/admin) |
| Kibana | http://localhost:5601 |
| Elasticsearch | http://localhost:9200 |
| Alertmanager | http://localhost:9093 |
| OTel Collector | localhost:4317 (gRPC), localhost:4318 (HTTP) |

Stop with `make obs-down`.

## Environment Variables

All environment variables are pre-configured in `docker-compose.yml` with sensible defaults for local development:

- Auth is disabled (`AUTH_ENABLED=false`)
- AWS credentials are set to `test`/`test` for DynamoDB Local
- JWT secret is set to a development value
- CORS allows all origins

No `.env` file configuration is needed. If you want to customize values, you can set them in the `environment` section of each service in `docker-compose.yml`.

## Troubleshooting

### Port Already in Use

If a port is already in use by another process:

```bash
# Find what's using a port
lsof -i :8080  # or netstat -tlnp | grep 8080

# Stop everything and try again
make dev-down
make dev
```

### Containers Won't Start

```bash
# Check container logs
docker compose logs go-service
docker compose logs python-service

# Full reset
docker compose down -v --rmi local
make dev
```

### DynamoDB Table Issues

The table is created automatically by the `init-dynamodb` service. If it fails:

```bash
# Create manually
make create-table
```

### Build Failures

```bash
# Force rebuild without cache
docker compose build --no-cache
docker compose up -d
```

## Running Without Docker

If you need to run services individually for debugging:

1. Start DynamoDB Local:
   ```bash
   docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest
   make create-table
   ```

2. Start each service in separate terminals:
   ```bash
   # Go backend
   cd backend/go && cp .env.example .env && go run cmd/server/main.go

   # Python analytics
   cd backend/python && cp .env.example .env && pip install -r analytics_service/requirements.txt && uvicorn analytics_service.main:app --port 8081

   # Frontend
   cd frontend && cp .env.example .env && npm install && npm start

   # Firmware simulator
   cd firmware && cp .env.example .env && go run cmd/firmware-sim/main.go
   ```
