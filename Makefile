.PHONY: up down migrate seed test lint clean build help

# Docker commands
up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

logs: ## View logs
	docker-compose logs -f

status: ## Show status of all containers
	docker-compose ps

# Database commands
migrate-up: ## Run database migrations
	migrate -path migrations -database "mysql://airline_user:airline_pass@tcp(localhost:3306)/airline_booking" up

migrate-down: ## Rollback database migrations
	migrate -path migrations -database "mysql://airline_user:airline_pass@tcp(localhost:3306)/airline_booking" down

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	migrate create -ext sql -dir migrations $(name)

# Seed data
seed: ## Seed database and Elasticsearch with sample data
	@echo "Seeding database and Elasticsearch..."
	docker-compose exec app go run ./cmd/seeder/main.go
	@echo "==> Seeding completed!"

seed-sql: ## Seed database using SQL file (alternative method)
	@echo "Seeding database with SQL file..."
	docker-compose exec -T mysql mysql -u root -prootpass airline_booking < seed_data.sql
	@echo "==> Database seeded with SQL!"

# Development
dev: ## Run in development mode
	docker-compose exec app go run cmd/api/main.go

build: ## Build the application
	docker-compose exec app go build -o bin/airline-api cmd/api/main.go

# Testing
test: ## Run tests (simplified without external dependencies)
	@echo "Running unit tests..."
	docker build --target builder -t airline-test-temp .
	docker run --rm airline-test-temp go test -v ./internal/... -short

test-race: ## Run tests with race detection  
	@echo "Running tests with race detection..."
	docker build --target builder -t airline-test-temp .
	docker run --rm -e CGO_ENABLED=1 airline-test-temp go test -race -v ./internal/... -short

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	docker build --target builder -t airline-test-temp .
	docker run --rm airline-test-temp sh -c "go test -coverprofile=coverage.out ./internal/... -short && go tool cover -func=coverage.out"

# Code quality
lint: ## Run linter
	golangci-lint run

# API testing examples
test-search: ## Test flight search API
	@echo "Testing flight search..."
	curl -s "http://localhost:8080/api/v1/flights/search?origin=JFK&destination=LAX&date=2025-08-30" | jq .

test-create-flight: ## Test flight creation API
	@echo "Testing flight creation..."
	curl -s -X POST "http://localhost:8080/api/v1/flights" \
		-H "Content-Type: application/json" \
		-d '{"origin": "BOS", "destination": "SFO", "departure_time": "2025-08-31T08:00:00Z", "arrival_time": "2025-08-31T14:30:00Z", "airline": "UA", "aircraft": "Boeing 737", "fare_class": "economy", "base_price": 399.99, "seat_config": {"economy_rows": 25, "business_rows": 3, "first_class_rows": 1, "seats_per_row": 6}}' | jq .

test-seats: ## Test seat availability API
	@echo "Testing seat availability..."
	curl -s "http://localhost:8080/api/v1/flights/1/seats" | jq .

test-hold: ## Test creating a hold
	@echo "Testing hold creation..."
	curl -s -X POST "http://localhost:8080/api/v1/holds" \
		-H "Content-Type: application/json" \
		-H "User-ID: testuser" \
		-d '{"flight_id": 1, "seat_no": "25A"}' | jq .

# URLs for quick access
urls: ## Show important URLs
	@echo "==> Important URLs:"
	@echo "API Health:       http://localhost:8080/api/v1/health"
	@echo "phpMyAdmin:       http://localhost:8081"
	@echo "Kibana:           http://localhost:5601"
	@echo "Elasticsearch:    http://localhost:9200"

# Database operations
db-shell: ## Open MySQL shell
	docker-compose exec mysql mysql -u root -prootpass airline_booking

# Elasticsearch operations
es-health: ## Check Elasticsearch health
	curl -s http://localhost:9200/_cluster/health | jq .

# Installation and setup
install: ## Complete project setup from scratch
	@echo "==> Installing Airline Booking API from scratch..."
	@echo "Step 1/6: Stopping any running containers..."
	-docker-compose down -v 2>/dev/null || true
	@echo "Step 2/6: Building application..."
	docker-compose build
	@echo "Step 3/6: Starting services..."
	docker-compose up -d
	@echo "Step 4/6: Waiting for services to be ready..."
	@sleep 45
	@echo "Step 5/6: Running database migrations..."
	@./scripts/wait-for-db.sh
	@echo "Step 6/6: Seeding database and Elasticsearch..."
	@./scripts/seed-all.sh
	@echo "==> Installation completed successfully!"
	@echo ""
	$(MAKE) urls

install-quick: ## Quick setup (assumes services are running)
	@echo "==> Quick setup for Airline Booking API..."
	@echo "Seeding database and Elasticsearch..."
	@./scripts/seed-all.sh
	@echo "==> Quick setup completed!"
	$(MAKE) urls

format: ## Format code
	docker-compose exec app go fmt ./...
	docker-compose exec app goimports -w .

# SQLC
sqlc-generate: ## Generate SQLC code
	docker-compose exec app sqlc generate

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.out coverage.html
	docker-compose down -v

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
