#!/bin/bash

# Airline Booking API - Test Examples
# Make sure the API is running on localhost:8080

API_BASE="http://localhost:8080/api/v1"
HEALTH_URL="http://localhost:8080/health"

echo "🚀 Airline Booking API - Test Examples"
echo "======================================="

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

# Check if API is running
print_step "Checking API health..."
if curl -s "$HEALTH_URL" > /dev/null; then
    print_success "API is running!"
else
    print_error "API is not running. Please start it with: make up"
    exit 1
fi

echo ""

# 1. Health Check
print_step "1. Health Check"
echo "Request: GET $HEALTH_URL"
curl -s "$HEALTH_URL" | jq '.'
echo ""

# 2. Search for flights
print_step "2. Search for flights (JFK to LAX)"
echo "Request: GET $API_BASE/flights/search?origin=JFK&destination=LAX&date=2025-08-30"
curl -s -G "$API_BASE/flights/search" \
  -d "origin=JFK" \
  -d "destination=LAX" \
  -d "date=2025-08-30" | jq '.'
echo ""

# 3. Get flight seat availability
print_step "3. Get flight seat availability (Flight ID: 1)"
echo "Request: GET $API_BASE/flights/1/seats"
curl -s "$API_BASE/flights/1/seats" | jq '.[0:5]' # Show first 5 seats
echo "... (showing first 5 seats only)"
echo ""

# 4. Create a seat hold
print_step "4. Create a seat hold (Flight 1, Seat 12A)"
USER_ID="user_$(date +%s)" # Unique user ID
IDEMPOTENCY_KEY=$(uuidgen 2>/dev/null || echo "test-key-$(date +%s)")
echo "User-ID: $USER_ID"
echo "Idempotency-Key: $IDEMPOTENCY_KEY"
echo "Request: POST $API_BASE/holds"

HOLD_RESPONSE=$(curl -s -X POST "$API_BASE/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: $USER_ID" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d '{
    "flight_id": 1,
    "seat_no": "12A"
  }')

echo "$HOLD_RESPONSE" | jq '.'

# Check if hold was successful
if echo "$HOLD_RESPONSE" | jq -e '.expires_at' > /dev/null; then
    print_success "Hold created successfully!"
    EXPIRES_AT=$(echo "$HOLD_RESPONSE" | jq -r '.expires_at')
    print_warning "Hold expires at: $EXPIRES_AT"
else
    print_error "Failed to create hold. Seat might be already taken."
fi
echo ""

# 5. Try to create the same hold (test idempotency)
print_step "5. Test idempotency - same request with same key"
echo "Request: POST $API_BASE/holds (same idempotency key)"
curl -s -X POST "$API_BASE/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: $USER_ID" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d '{
    "flight_id": 1,
    "seat_no": "12A"
  }' | jq '.'
echo ""

# 6. Try concurrent holds (different users, same seat)
print_step "6. Test concurrency - multiple users trying same seat"
echo "Starting 3 concurrent requests for seat 15B..."

for i in {1..3}; do
    {
        USER="concurrent_user_$i"
        RESULT=$(curl -s -X POST "$API_BASE/holds" \
          -H "Content-Type: application/json" \
          -H "User-ID: $USER" \
          -d '{
            "flight_id": 1,
            "seat_no": "15B"
          }')
        
        if echo "$RESULT" | jq -e '.expires_at' > /dev/null; then
            echo "✅ User $i: SUCCESS"
        else
            ERROR_MSG=$(echo "$RESULT" | jq -r '.message // "Unknown error"')
            echo "❌ User $i: FAILED - $ERROR_MSG"
        fi
    } &
done

wait # Wait for all background jobs to complete
echo ""

# 7. Confirm ticket (if we have a valid hold)
if echo "$HOLD_RESPONSE" | jq -e '.expires_at' > /dev/null; then
    print_step "7. Confirm ticket purchase"
    PAYMENT_REF="payment_$(date +%s)"
    TICKET_IDEMPOTENCY_KEY=$(uuidgen 2>/dev/null || echo "ticket-key-$(date +%s)")
    
    echo "Request: POST $API_BASE/tickets/confirm"
    echo "Payment Reference: $PAYMENT_REF"
    
    TICKET_RESPONSE=$(curl -s -X POST "$API_BASE/tickets/confirm" \
      -H "Content-Type: application/json" \
      -H "User-ID: $USER_ID" \
      -H "Idempotency-Key: $TICKET_IDEMPOTENCY_KEY" \
      -d "{
        \"flight_id\": 1,
        \"seat_no\": \"12A\",
        \"payment_ref\": \"$PAYMENT_REF\"
      }")
    
    echo "$TICKET_RESPONSE" | jq '.'
    
    if echo "$TICKET_RESPONSE" | jq -e '.pnr_code' > /dev/null; then
        PNR_CODE=$(echo "$TICKET_RESPONSE" | jq -r '.pnr_code')
        print_success "Ticket confirmed! PNR: $PNR_CODE"
    else
        print_error "Failed to confirm ticket"
    fi
    echo ""
fi

# 8. Try to hold an already sold seat
print_step "8. Try to hold an already sold seat (should fail)"
echo "Request: POST $API_BASE/holds (seat 12A - should be sold now)"
curl -s -X POST "$API_BASE/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: another_user_$(date +%s)" \
  -d '{
    "flight_id": 1,
    "seat_no": "12A"
  }' | jq '.'
echo ""

# 9. Release a hold
print_step "9. Create and release a hold"
TEMP_USER="temp_user_$(date +%s)"

# Create hold
echo "Creating hold for seat 20C..."
TEMP_HOLD=$(curl -s -X POST "$API_BASE/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: $TEMP_USER" \
  -d '{
    "flight_id": 1,
    "seat_no": "20C"
  }')

if echo "$TEMP_HOLD" | jq -e '.expires_at' > /dev/null; then
    print_success "Hold created for seat 20C"
    
    # Release hold
    echo "Releasing hold..."
    curl -s -X DELETE "$API_BASE/holds/1/20C" \
      -H "User-ID: $TEMP_USER"
    
    echo "Hold released (no content response expected)"
else
    print_error "Failed to create temporary hold"
fi
echo ""

# 10. Search with filters
print_step "10. Search flights with filters"
echo "Request: GET $API_BASE/flights/search?origin=JFK&destination=MIA&date=2025-08-30&fare_class=business"
curl -s -G "$API_BASE/flights/search" \
  -d "origin=JFK" \
  -d "destination=MIA" \
  -d "date=2025-08-30" \
  -d "fare_class=business" | jq '.'
echo ""

# 11. Check updated seat availability
print_step "11. Check updated seat availability"
echo "Request: GET $API_BASE/flights/1/seats (showing seats 12A-12F)"
curl -s "$API_BASE/flights/1/seats" | jq '.[] | select(.seat_no | startswith("12"))'
echo ""

print_step "🎉 Test sequence completed!"
echo ""
print_warning "Key observations:"
echo "- Only one user can hold the same seat at a time"
echo "- Idempotency keys prevent duplicate operations"
echo "- Sold seats cannot be held again"
echo "- Holds can be manually released"
echo "- All operations are logged and auditable"

echo ""
print_step "📊 To monitor the system:"
echo "- Check logs: docker-compose logs app"
echo "- View Swagger UI: http://localhost:8080/docs/"
echo "- Access Kibana: http://localhost:5601 (optional)"
