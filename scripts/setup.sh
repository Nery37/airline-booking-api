#!/bin/bash

# Airline Booking API - Setup Script
# This script helps set up the project for the first time

echo "🚀 Airline Booking API - Setup"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check prerequisites
print_step "Checking prerequisites..."

# Check Docker
if command -v docker &> /dev/null; then
    print_success "Docker is installed"
else
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check Docker Compose
if command -v docker-compose &> /dev/null; then
    print_success "Docker Compose is installed"
else
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if .env exists, if not create from example
if [ ! -f .env ]; then
    print_step "Creating .env file from .env.example..."
    cp .env.example .env
    print_success ".env file created"
else
    print_warning ".env file already exists"
fi

echo ""

# Pull required Docker images
print_step "Pulling required Docker images..."
docker-compose pull

echo ""

# Start services
print_step "Starting services..."
docker-compose up -d

# Wait for services to be ready
print_step "Waiting for services to be ready..."
echo "This may take a few minutes for the first time..."

# Wait for MySQL
print_step "Waiting for MySQL..."
timeout=60
while ! docker-compose exec -T mysql mysqladmin ping -h localhost --silent; do
    if [ $timeout -le 0 ]; then
        print_error "MySQL failed to start within timeout"
        exit 1
    fi
    sleep 2
    timeout=$((timeout-2))
done
print_success "MySQL is ready"

# Wait for Elasticsearch
print_step "Waiting for Elasticsearch..."
timeout=60
while ! curl -s http://localhost:9200/_cluster/health > /dev/null; do
    if [ $timeout -le 0 ]; then
        print_error "Elasticsearch failed to start within timeout"
        exit 1
    fi
    sleep 2
    timeout=$((timeout-2))
done
print_success "Elasticsearch is ready"

echo ""

# Check if Go is available for migrations and seeding
if command -v go &> /dev/null; then
    print_step "Go is available, running migrations and seeding..."
    
    # Download dependencies
    print_step "Downloading Go dependencies..."
    go mod download
    
    # Install golang-migrate if not exists
    if ! command -v migrate &> /dev/null; then
        print_step "Installing golang-migrate..."
        go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    fi
    
    # Run migrations
    print_step "Running database migrations..."
    if make migrate-up; then
        print_success "Migrations completed"
    else
        print_warning "Migrations failed, trying manual approach..."
        # Alternative migration approach
        docker-compose exec -T mysql mysql -u airline_user -pairline_pass airline_booking < migrations/000001_initial_schema.up.sql
    fi
    
    # Seed database
    print_step "Seeding database..."
    if timeout 30s go run cmd/seed/main.go; then
        print_success "Database seeded"
    else
        print_warning "Database seeding failed or timed out"
    fi
    
    # Seed Elasticsearch
    print_step "Seeding Elasticsearch..."
    if timeout 30s go run cmd/es-seed/main.go; then
        print_success "Elasticsearch seeded"
    else
        print_warning "Elasticsearch seeding failed or timed out"
    fi
    
else
    print_warning "Go is not installed. Skipping migrations and seeding."
    print_warning "You can run them manually later with:"
    echo "  make migrate-up"
    echo "  make seed"
    echo "  make es-seed"
fi

echo ""

# Check if API is responding
print_step "Checking API health..."
timeout=30
while ! curl -s http://localhost:8080/health > /dev/null; do
    if [ $timeout -le 0 ]; then
        print_warning "API is not responding yet. It may still be starting up."
        break
    fi
    sleep 2
    timeout=$((timeout-2))
done

if curl -s http://localhost:8080/health > /dev/null; then
    print_success "API is responding!"
else
    print_warning "API is not responding yet. Check logs with: docker-compose logs app"
fi

echo ""
print_step "🎉 Setup completed!"
echo ""
print_success "Services available:"
echo "  🌐 API: http://localhost:8080"
echo "  📖 Swagger UI: http://localhost:8080/docs/"
echo "  🔍 Elasticsearch: http://localhost:9200"
echo "  📊 Kibana: http://localhost:5601"
echo ""
print_warning "Next steps:"
echo "  1. Test the API: bash scripts/test_api.sh"
echo "  2. View logs: docker-compose logs -f app"
echo "  3. Stop services: make down"
echo ""
print_step "📚 Documentation:"
echo "  • README.md - Complete documentation"
echo "  • docs/openapi.yaml - API specification"
echo "  • Run 'make help' for all available commands"
