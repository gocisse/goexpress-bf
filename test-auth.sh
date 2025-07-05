#!/bin/bash

# GoExpress Authentication Test Script
# This script specifically tests authentication flows

BASE_URL="http://localhost:8080"

echo "🔐 Testing GoExpress Authentication"
echo "=================================="

# Test 1: Login with default admin
echo "1. Testing Admin Login"
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@goexpress.com",
    "password": "goexpress123"
  }')

echo "Response: $ADMIN_LOGIN_RESPONSE"

# Extract token
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ADMIN_TOKEN" ]; then
    echo "✅ Admin login successful"
    echo "Token: $ADMIN_TOKEN"
else
    echo "❌ Admin login failed"
fi

echo ""

# Test 2: Register new client
echo "2. Testing Client Registration"
CLIENT_REG_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Client",
    "email": "testclient@goexpress.com",
    "password": "client123",
    "role": "client"
  }')

echo "Response: $CLIENT_REG_RESPONSE"

# Extract client token
CLIENT_TOKEN=$(echo "$CLIENT_REG_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$CLIENT_TOKEN" ]; then
    echo "✅ Client registration successful"
    echo "Token: $CLIENT_TOKEN"
else
    echo "❌ Client registration failed"
fi

echo ""

# Test 3: Test protected endpoint with admin token
echo "3. Testing Protected Endpoint (Admin)"
if [ -n "$ADMIN_TOKEN" ]; then
    curl -X GET "$BASE_URL/api/shipments" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -w "\nStatus: %{http_code}\n"
else
    echo "No admin token available"
fi

echo ""

# Test 4: Test protected endpoint with client token
echo "4. Testing Protected Endpoint (Client)"
if [ -n "$CLIENT_TOKEN" ]; then
    curl -X GET "$BASE_URL/api/shipments" \
      -H "Authorization: Bearer $CLIENT_TOKEN" \
      -w "\nStatus: %{http_code}\n"
else
    echo "No client token available"
fi

echo ""

# Test 5: Test admin-only endpoint with client token (should fail)
echo "5. Testing Admin Endpoint with Client Token (Should Fail)"
if [ -n "$CLIENT_TOKEN" ]; then
    curl -X POST "$BASE_URL/api/zones" \
      -H "Authorization: Bearer $CLIENT_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Test Zone",
        "price_per_kg": 5.00
      }' \
      -w "\nStatus: %{http_code}\n"
else
    echo "No client token available"
fi

echo ""
echo "🎉 Authentication tests completed!"
