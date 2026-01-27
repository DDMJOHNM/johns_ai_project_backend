#!/bin/bash
# Script to test authentication endpoints

set -e

# Configuration
BASE_URL="${API_URL:-http://localhost:8080}"
RANDOM_SUFFIX=$(date +%s)

echo "=== Testing Authentication System ==="
echo "Base URL: $BASE_URL"
echo ""

# Test data
USERNAME="testuser_${RANDOM_SUFFIX}"
EMAIL="test_${RANDOM_SUFFIX}@example.com"
PASSWORD="TestPassword123!"
FIRST_NAME="Test"
LAST_NAME="User"

echo "1. Testing User Registration..."
echo "   Username: $USERNAME"
echo "   Email: $EMAIL"

REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"email\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\",
    \"first_name\": \"${FIRST_NAME}\",
    \"last_name\": \"${LAST_NAME}\"
  }")

echo "$REGISTER_RESPONSE" | jq '.'

# Extract token from registration
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Registration failed - no token received"
  exit 1
fi

echo "✅ Registration successful"
echo "   Token: ${TOKEN:0:20}..."
echo ""

echo "2. Testing /api/auth/me endpoint..."
ME_RESPONSE=$(curl -s "${BASE_URL}/api/auth/me" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$ME_RESPONSE" | jq '.'

USER_ID=$(echo "$ME_RESPONSE" | jq -r '.id')
if [ "$USER_ID" == "null" ] || [ -z "$USER_ID" ]; then
  echo "❌ /api/auth/me failed"
  exit 1
fi

echo "✅ /api/auth/me successful"
echo ""

echo "3. Testing Login with Username..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"login\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "$LOGIN_RESPONSE" | jq '.'

LOGIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
if [ "$LOGIN_TOKEN" == "null" ] || [ -z "$LOGIN_TOKEN" ]; then
  echo "❌ Login with username failed"
  exit 1
fi

echo "✅ Login with username successful"
echo ""

echo "4. Testing Login with Email..."
LOGIN_EMAIL_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"login\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "$LOGIN_EMAIL_RESPONSE" | jq '.'

LOGIN_EMAIL_TOKEN=$(echo "$LOGIN_EMAIL_RESPONSE" | jq -r '.token')
if [ "$LOGIN_EMAIL_TOKEN" == "null" ] || [ -z "$LOGIN_EMAIL_TOKEN" ]; then
  echo "❌ Login with email failed"
  exit 1
fi

echo "✅ Login with email successful"
echo ""

echo "5. Testing Protected Endpoint (GET /api/clients)..."
CLIENTS_RESPONSE=$(curl -s "${BASE_URL}/api/clients" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$CLIENTS_RESPONSE" | jq '.'

if echo "$CLIENTS_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo "❌ Protected endpoint failed"
  exit 1
fi

echo "✅ Protected endpoint accessible with valid token"
echo ""

echo "6. Testing Protected Endpoint without Token (should fail)..."
UNAUTH_RESPONSE=$(curl -s "${BASE_URL}/api/clients")

echo "$UNAUTH_RESPONSE" | jq '.'

if echo "$UNAUTH_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
  echo "✅ Unauthorized request correctly blocked"
else
  echo "❌ Unauthorized request was not blocked (security issue!)"
  exit 1
fi

echo ""
echo "════════════════════════════════════════════"
echo "✅ All authentication tests passed!"
echo "════════════════════════════════════════════"
echo ""
echo "Test credentials:"
echo "  Username: $USERNAME"
echo "  Email: $EMAIL"
echo "  Password: $PASSWORD"
echo "  Token: $TOKEN"
echo ""
echo "You can use these to test in Postman or other tools."

