#!/bin/bash

# GoExpress Delivery Management API Test Script
# This script tests all the API endpoints

BASE_URL="http://localhost:8080"
ADMIN_TOKEN=""
CLIENT_TOKEN=""
DRIVER_TOKEN=""
TRACKING_NUMBER=""

echo "🚀 Testing GoExpress Delivery Management API"
echo "============================================="

# Test 1: Health Check
echo "1. Health Check"
curl -X GET "$BASE_URL/health" -w "\nStatus: %{http_code}\n\n"

# Test 2: Root Endpoint
echo "2. Root Endpoint"
curl -X GET "$BASE_URL/" -w "\nStatus: %{http_code}\n\n"

# Test 3: Login Admin (instead of register, since admin exists)
echo "3. Login Admin User"
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@goexpress.com",
    "password": "goexpress123"
  }' -w "\nStatus: %{http_code}")

echo "$ADMIN_RESPONSE"
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Admin Token: $ADMIN_TOKEN"
echo

# Test 4: Register Client User
echo "4. Register Client User"
CLIENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GoExpress Client",
    "email": "client@goexpress.com",
    "password": "client123",
    "role": "client"
  }' -w "\nStatus: %{http_code}")

echo "$CLIENT_RESPONSE"
CLIENT_TOKEN=$(echo "$CLIENT_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Client Token: $CLIENT_TOKEN"
echo

# Test 5: Register Driver User
echo "5. Register Driver User"
DRIVER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GoExpress Driver",
    "email": "driver@goexpress.com",
    "password": "driver123",
    "role": "driver"
  }' -w "\nStatus: %{http_code}")

echo "$DRIVER_RESPONSE"
DRIVER_TOKEN=$(echo "$DRIVER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Driver Token: $DRIVER_TOKEN"
echo

# Test 6: Login Test (Verify admin login again)
echo "6. Login Test (Admin Verification)"
curl -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@goexpress.com",
    "password": "goexpress123"
  }' -w "\nStatus: %{http_code}\n\n"

# Test 7: Get All Zones (Public)
echo "7. Get All GoExpress Zones (Public)"
curl -X GET "$BASE_URL/api/zones" -w "\nStatus: %{http_code}\n\n"

# Test 8: Create New Zone (Admin Only)
echo "8. Create New Zone (Admin Only)"
if [ -n "$ADMIN_TOKEN" ]; then
  curl -X POST "$BASE_URL/api/zones" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Premium Express",
      "price_per_kg": 20.00
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No admin token available"
  echo "Status: N/A"
  echo
fi

# Test 9: Get Shipping Quote (Public)
echo "9. Get GoExpress Shipping Quote (Public)"
curl -X POST "$BASE_URL/api/quote" \
  -H "Content-Type: application/json" \
  -d '{
    "weight": 2.5,
    "zone_id": 1
  }' -w "\nStatus: %{http_code}\n\n"

# Test 10: Create Shipment (Client)
echo "10. Create Shipment (Client)"
if [ -n "$CLIENT_TOKEN" ]; then
  SHIPMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $CLIENT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "origin": "Mumbai, India",
      "destination": "Delhi, India",
      "weight": 3.5,
      "zone_id": 2
    }' -w "\nStatus: %{http_code}")

  echo "$SHIPMENT_RESPONSE"
  TRACKING_NUMBER=$(echo "$SHIPMENT_RESPONSE" | grep -o '"tracking_number":"[^"]*"' | cut -d'"' -f4)
  echo "GoExpress Tracking Number: $TRACKING_NUMBER"
  echo
else
  echo "No client token available"
  echo "Status: N/A"
  echo
fi

# Test 11: Get All Shipments (Client - should see only their own)
echo "11. Get All Shipments (Client - own shipments only)"
if [ -n "$CLIENT_TOKEN" ]; then
  curl -X GET "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $CLIENT_TOKEN" \
    -w "\nStatus: %{http_code}\n\n"
else
  echo "No client token available"
  echo "Status: N/A"
  echo
fi

# Test 12: Get All Shipments (Admin - should see all)
echo "12. Get All Shipments (Admin - all shipments)"
if [ -n "$ADMIN_TOKEN" ]; then
  curl -X GET "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -w "\nStatus: %{http_code}\n\n"
else
  echo "No admin token available"
  echo "Status: N/A"
  echo
fi

# Test 13: Track Shipment (Public)
echo "13. Track GoExpress Shipment by Tracking Number (Public)"
if [ -n "$TRACKING_NUMBER" ]; then
  curl -X GET "$BASE_URL/api/shipments/$TRACKING_NUMBER" \
    -w "\nStatus: %{http_code}\n\n"
else
  echo "No tracking number available"
fi

# Test 14: Update Shipment Status (Admin)
echo "14. Update Shipment Status (Admin)"
if [ -n "$ADMIN_TOKEN" ] && [ -n "$TRACKING_NUMBER" ]; then
  # Use shipment ID 1 (first shipment created)
  SHIPMENT_ID=1
  
  curl -X PUT "$BASE_URL/api/shipments/$SHIPMENT_ID/status" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "status": "in_transit",
      "location": "GoExpress Hub - Pune"
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No admin token or shipment available to update"
fi

# Test 15: Unauthorized Access Test
echo "15. Unauthorized Access Test (No Token)"
curl -X GET "$BASE_URL/api/shipments" \
  -w "\nStatus: %{http_code}\n\n"

# Test 16: Forbidden Access Test (Client trying to access admin endpoint)
echo "16. Forbidden Access Test (Client trying to create zone)"
if [ -n "$CLIENT_TOKEN" ]; then
  curl -X POST "$BASE_URL/api/zones" \
    -H "Authorization: Bearer $CLIENT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Test Zone",
      "price_per_kg": 5.00
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No client token available"
  echo "Status: N/A"
  echo
fi

# Test 17: Invalid Token Test
echo "17. Invalid Token Test"
curl -X GET "$BASE_URL/api/shipments" \
  -H "Authorization: Bearer invalid_token_here" \
  -w "\nStatus: %{http_code}\n\n"

# Test 18: Update Zone (Admin Only)
echo "18. Update Zone (Admin Only)"
if [ -n "$ADMIN_TOKEN" ]; then
  curl -X PUT "$BASE_URL/api/zones/1" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Local Express Premium",
      "price_per_kg": 4.00
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No admin token available"
  echo "Status: N/A"
  echo
fi

# Test 19: Track Non-existent Shipment
echo "19. Track Non-existent GoExpress Shipment"
curl -X GET "$BASE_URL/api/shipments/GEX00000000" \
  -w "\nStatus: %{http_code}\n\n"

# Test 20: Create Shipment with Invalid Data
echo "20. Create Shipment with Invalid Data"
if [ -n "$CLIENT_TOKEN" ]; then
  curl -X POST "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $CLIENT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "origin": "",
      "destination": "Delhi, India",
      "weight": -1,
      "zone_id": 999
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No client token available"
  echo "Status: N/A"
  echo
fi

# Test 21: Quote with Invalid Data
echo "21. Quote with Invalid Data"
curl -X POST "$BASE_URL/api/quote" \
  -H "Content-Type: application/json" \
  -d '{
    "weight": 0,
    "zone_id": 999
  }' -w "\nStatus: %{http_code}\n\n"

# Test 22: Create Another Shipment for Testing
echo "22. Create Another Shipment (Different Route)"
if [ -n "$CLIENT_TOKEN" ]; then
  curl -X POST "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $CLIENT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "origin": "Bangalore, India",
      "destination": "Chennai, India",
      "weight": 1.2,
      "zone_id": 1
    }' -w "\nStatus: %{http_code}\n\n"
else
  echo "No client token available"
  echo "Status: N/A"
  echo
fi

# Test 23: Test Driver Access (Driver viewing their shipments)
echo "23. Test Driver Access (Driver shipments)"
if [ -n "$DRIVER_TOKEN" ]; then
  curl -X GET "$BASE_URL/api/shipments" \
    -H "Authorization: Bearer $DRIVER_TOKEN" \
    -w "\nStatus: %{http_code}\n\n"
else
  echo "No driver token available"
  echo "Status: N/A"
  echo
fi

echo "🎉 All GoExpress API tests completed!"
echo "======================================"
echo "Check the responses above for any issues."
echo ""
echo "Expected behavior:"
echo "- Status 200/201 for successful operations"
echo "- Status 401 for unauthorized access"
echo "- Status 403 for forbidden access"
echo "- Status 400 for bad requests"
echo "- Status 404 for not found"
echo "- Status 409 for conflicts"
echo ""
echo "🌐 Access Swagger docs: http://localhost:8080/swagger/index.html"
echo "🏥 Health check: http://localhost:8080/health"
echo "📊 Default admin: admin@goexpress.com / goexpress123"
echo ""
echo "🚚 GoExpress - Fast, Reliable, Tracked! ✨"
