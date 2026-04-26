# Sentinel Device Manager - Makefile
# Common commands for building, testing, and running the platform

.PHONY: help dev dev-down test test-go test-python test-firmware test-all \
        test-coverage test-integration test-e2e test-postman test-playwright \
        docker-up docker-down swagger godoc lint lint-go lint-python \
        clean create-table trivy-scan vuln-check \
        obs-up obs-down obs-logs obs-grafana obs-prometheus obs-kibana

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Local Development ---

dev: ## Start all services with Docker Compose
	docker-compose up -d
	@echo "Services running:"
	@echo "  Frontend:   http://localhost:3000"
	@echo "  Go API:     http://localhost:8080"
	@echo "  gRPC:       localhost:9090"
	@echo "  Python API: http://localhost:8081"
	@echo "  DynamoDB:   http://localhost:8000/shell"
	@echo "  Swagger:    http://localhost:8080/swagger.json"

dev-down: ## Stop all services
	docker-compose down

create-table: ## Create DynamoDB table for local development
	aws dynamodb create-table \
		--table-name sentinel-devices \
		--attribute-definitions AttributeName=device_id,AttributeType=S \
		--key-schema AttributeName=device_id,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST \
		--endpoint-url http://localhost:8000 \
		--region us-east-1

# --- Testing ---

test: test-go test-python test-firmware ## Run all unit tests

test-go: ## Run Go backend tests
	cd backend/go && go test ./... -v -race -count=1

test-python: ## Run Python analytics tests
	cd backend/python && pytest analytics_service/tests/ -v

test-firmware: ## Run firmware simulator tests
	cd firmware && go test ./... -v

test-coverage: ## Run Go tests with coverage report (enforces 70%)
	cd backend/go && go test ./... -race -coverprofile=coverage.out -covermode=atomic -count=1 && \
		go tool cover -func=coverage.out && \
		COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//') && \
		echo "Total coverage: $${COVERAGE}%" && \
		if [ "$$(echo "$$COVERAGE < 70" | bc -l)" -eq 1 ]; then echo "ERROR: Coverage below 70%"; exit 1; fi

test-integration: ## Run integration tests (requires running services or in-memory)
	cd tests/integration && go test ./... -v -count=1 -timeout 120s

test-e2e: ## Run end-to-end tests (requires running services)
	cd tests/e2e && go test ./... -v -timeout 120s

test-playwright: ## Run Playwright browser tests (Chromium only)
	cd tests/playwright && npm ci && npx playwright install --with-deps chromium && npx playwright test --project=chromium

test-postman: ## Run Postman Newman API contract tests
	newman run tests/postman/sentinel-api-collection.json \
		-e tests/postman/sentinel-api-environment.json \
		--reporters cli,json \
		--reporter-json-export tests/postman/newman-report.json

test-all: test test-coverage test-integration test-postman ## Run all tests

# --- Linting ---

lint: lint-go lint-python ## Run all linters

lint-go: ## Run Go linters (golangci-lint)
	cd backend/go && golangci-lint run ./... --timeout=5m
	cd firmware && golangci-lint run ./... --timeout=5m

lint-python: ## Run Python linter
	cd backend/python && flake8 analytics_service/ --max-line-length=120 --count --statistics

# --- Security ---

trivy-scan: ## Run Trivy vulnerability scan on source code
	trivy fs --severity CRITICAL,HIGH --ignorefile .trivyignore backend/go backend/python frontend --skip-dirs frontend/node_modules

vuln-check: ## Run Go vulnerability check
	cd backend/go && govulncheck ./...
	cd firmware && govulncheck ./...

# --- Documentation ---

swagger: ## View Swagger spec
	@curl -s http://localhost:8080/swagger.json | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/swagger.json

godoc: ## Start Go documentation server
	@echo "Open http://localhost:6060/pkg/github.com/sentinel-device-manager/backend/go/"
	cd backend/go && godoc -http=:6060

# --- Docker ---

docker-up: ## Build and start all Docker services
	docker-compose up -d --build

docker-down: ## Stop and remove all Docker resources
	docker-compose down -v

docker-logs: ## Tail logs from all services
	docker-compose logs -f

# --- Cleanup ---

clean: ## Remove build artifacts
	rm -rf backend/go/bin/ backend/go/coverage.out
	find . -name "*.test" -delete
	find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true
	rm -rf tests/playwright/node_modules tests/playwright/playwright-report
	rm -f tests/postman/newman-report.json

# --- Observability ---

obs-up: ## Start observability stack (Prometheus, Grafana, ELK, OTel)
	docker compose -f docker-compose.yml -f observability/docker-compose.observability.yml up -d
	@echo "Observability services running:"
	@echo "  Prometheus:    http://localhost:9091"
	@echo "  Grafana:       http://localhost:3001 (admin/admin)"
	@echo "  Kibana:        http://localhost:5601"
	@echo "  Elasticsearch: http://localhost:9200"
	@echo "  Alertmanager:  http://localhost:9093"
	@echo "  OTel Collector: localhost:4317 (gRPC), localhost:4318 (HTTP)"

obs-down: ## Stop observability stack
	docker compose -f docker-compose.yml -f observability/docker-compose.observability.yml down

obs-logs: ## Tail logs from observability services
	docker compose -f docker-compose.yml -f observability/docker-compose.observability.yml logs -f prometheus grafana elasticsearch logstash kibana otel-collector alertmanager

obs-grafana: ## Open Grafana dashboard
	@echo "Opening Grafana at http://localhost:3001"
	@open http://localhost:3001 2>/dev/null || xdg-open http://localhost:3001 2>/dev/null || echo "Open http://localhost:3001 in your browser"

obs-prometheus: ## Open Prometheus UI
	@echo "Opening Prometheus at http://localhost:9091"
	@open http://localhost:9091 2>/dev/null || xdg-open http://localhost:9091 2>/dev/null || echo "Open http://localhost:9091 in your browser"

obs-kibana: ## Open Kibana dashboard
	@echo "Opening Kibana at http://localhost:5601"
	@open http://localhost:5601 2>/dev/null || xdg-open http://localhost:5601 2>/dev/null || echo "Open http://localhost:5601 in your browser"
