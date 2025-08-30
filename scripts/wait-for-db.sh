#!/bin/bash

echo "⏳ Waiting for database to be ready..."

max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if docker-compose exec -T mysql mysqladmin ping -h localhost -u root -prootpass --silent 2>/dev/null; then
        echo "✅ Database is ready!"
        
        # Wait a bit more for database to be fully initialized
        sleep 5
        
        # Check if our database exists, create if not
        if ! docker-compose exec -T mysql mysql -u root -prootpass -e "USE airline_booking;" 2>/dev/null; then
            echo "📋 Creating airline_booking database..."
            docker-compose exec -T mysql mysql -u root -prootpass -e "CREATE DATABASE IF NOT EXISTS airline_booking;" 2>/dev/null
            docker-compose exec -T mysql mysql -u root -prootpass -e "GRANT ALL PRIVILEGES ON airline_booking.* TO 'airline_user'@'%';" 2>/dev/null
            docker-compose exec -T mysql mysql -u root -prootpass -e "FLUSH PRIVILEGES;" 2>/dev/null
        fi
        
        exit 0
    fi
    
    echo "⏳ Attempt $attempt/$max_attempts: Database not ready yet, waiting..."
    sleep 2
    attempt=$((attempt + 1))
done

echo "❌ Database failed to start within expected time"
exit 1
