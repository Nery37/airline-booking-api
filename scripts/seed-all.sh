#!/bin/bash

echo "🌱 Starting comprehensive data seeding..."

# Function to wait for service
wait_for_service() {
    local service_name=$1
    local health_url=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$health_url" > /dev/null 2>&1; then
            echo "✅ $service_name is ready!"
            return 0
        fi
        
        echo "⏳ Attempt $attempt/$max_attempts: $service_name not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ $service_name failed to start within expected time"
    return 1
}

# Wait for Elasticsearch
wait_for_service "Elasticsearch" "http://localhost:9200/_cluster/health"

# Wait for application
wait_for_service "Application" "http://localhost:8080/api/v1/health"

echo "📋 Step 1/3: Seeding MySQL database..."
if docker-compose exec -T mysql mysql -u airline_user -pairline_pass airline_booking < seed_data.sql; then
    echo "✅ MySQL database seeded successfully!"
else
    echo "❌ Failed to seed MySQL database"
    exit 1
fi

echo "📋 Step 2/3: Creating Elasticsearch index..."
# Delete existing index if it exists
curl -s -X DELETE "http://localhost:9200/flights" > /dev/null 2>&1 || true

# Create the index with proper mapping
curl -s -X PUT "http://localhost:9200/flights" \
  -H "Content-Type: application/json" \
  -d '{
    "mappings": {
      "properties": {
        "id": {"type": "long"},
        "origin": {"type": "keyword"},
        "destination": {"type": "keyword"},
        "departure_time": {"type": "date"},
        "arrival_time": {"type": "date"},
        "airline": {"type": "keyword"},
        "aircraft": {"type": "text"},
        "fare_class": {"type": "keyword"},
        "base_price": {"type": "float"}
      }
    }
  }' > /dev/null

echo "📋 Step 3/3: Syncing flights to Elasticsearch..."
./sync_es.sh

echo "🎉 All seeding completed successfully!"
echo ""
echo "📊 Summary:"
echo "  - MySQL: $(docker-compose exec -T mysql mysql -u airline_user -pairline_pass airline_booking -s -N -e 'SELECT COUNT(*) FROM flights;' 2>/dev/null || echo 'N/A') flights"
echo "  - Elasticsearch: $(curl -s 'http://localhost:9200/flights/_count' | grep -o '"count":[0-9]*' | cut -d: -f2 || echo 'N/A') flights"
echo "  - Seats: $(docker-compose exec -T mysql mysql -u airline_user -pairline_pass airline_booking -s -N -e 'SELECT COUNT(*) FROM seats;' 2>/dev/null || echo 'N/A') seats"
