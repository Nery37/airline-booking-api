#!/bin/bash

# Script to sync MySQL flights to Elasticsearch

echo "Syncing flights from MySQL to Elasticsearch..."

# First, get all flights from MySQL
flights=$(docker exec -it airline_mysql mysql -u airline_user -pairline_pass airline_booking -s -N -e "SELECT id, origin, destination, departure_time, arrival_time, airline, aircraft, fare_class FROM flights;")

# Process each flight and index to Elasticsearch
while IFS=$'\t' read -r id origin dest dep_time arr_time airline aircraft fare_class; do
    if [ ! -z "$id" ]; then
        # Clean the data (remove any carriage returns)
        id=$(echo "$id" | tr -d '\r')
        origin=$(echo "$origin" | tr -d '\r')
        dest=$(echo "$dest" | tr -d '\r')
        dep_time=$(echo "$dep_time" | tr -d '\r')
        arr_time=$(echo "$arr_time" | tr -d '\r')
        airline=$(echo "$airline" | tr -d '\r')
        aircraft=$(echo "$aircraft" | tr -d '\r')
        fare_class=$(echo "$fare_class" | tr -d '\r')
        
        # Convert datetime format for Elasticsearch
        dep_time_es=$(echo "$dep_time" | sed 's/ /T/' | sed 's/$/Z/')
        arr_time_es=$(echo "$arr_time" | sed 's/ /T/' | sed 's/$/Z/')
        
        # Create the document
        doc="{
            \"id\": $id,
            \"origin\": \"$origin\",
            \"destination\": \"$dest\",
            \"departure_time\": \"$dep_time_es\",
            \"arrival_time\": \"$arr_time_es\",
            \"airline\": \"$airline\",
            \"aircraft\": \"$aircraft\",
            \"fare_class\": \"$fare_class\",
            \"base_price\": 299.99
        }"
        
        # Index to Elasticsearch
        curl -X POST "http://localhost:9200/flights/_doc/$id" \
             -H "Content-Type: application/json" \
             -d "$doc"
        
        echo "Indexed flight $id: $origin -> $dest"
    fi
done <<< "$flights"

echo "Elasticsearch sync completed!"
