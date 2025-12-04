#!/bin/bash

# API Testing Script for Mental Health Counselling Client Management System
# Usage: 
#   ./test-api.sh                    # Test local server
#   ./test-api.sh <api-gateway-url>  # Test API Gateway

if [ -z "$1" ]; then
  BASE_URL="${BASE_URL:-http://localhost:8080}"
  echo "Testing local server at: $BASE_URL"
else
  BASE_URL="$1"
  echo "Testing API Gateway at: $BASE_URL"
fi

echo "=========================================="
echo "Testing API Endpoints"
echo "Base URL: $BASE_URL"
echo "=========================================="
echo ""

# 1. Health Check
echo "1. Health Check"
echo "   GET /health"
curl -s -X GET "$BASE_URL/health" | jq '.' || curl -s -X GET "$BASE_URL/health"
echo ""
echo ""

# 2. Get All Clients
echo "2. Get All Clients"
echo "   GET /api/clients"
curl -s -X GET "$BASE_URL/api/clients" | jq '.' || curl -s -X GET "$BASE_URL/api/clients"
echo ""
echo ""

# 3. Get Active Clients
echo "3. Get Active Clients"
echo "   GET /api/clients/active"
curl -s -X GET "$BASE_URL/api/clients/active" | jq '.' || curl -s -X GET "$BASE_URL/api/clients/active"
echo ""
echo ""

# 4. Get Inactive Clients
echo "4. Get Inactive Clients"
echo "   GET /api/clients/inactive"
curl -s -X GET "$BASE_URL/api/clients/inactive" | jq '.' || curl -s -X GET "$BASE_URL/api/clients/inactive"
echo ""
echo ""

# 5. Get Client by ID (using seed data IDs)
echo "5. Get Client by ID"
echo "   GET /api/clients/client-001"
curl -s -X GET "$BASE_URL/api/clients/client-001" | jq '.' || curl -s -X GET "$BASE_URL/api/clients/client-001"
echo ""
echo ""

echo "6. Get Client by ID (another example)"
echo "   GET /api/clients/client-004"
curl -s -X GET "$BASE_URL/api/clients/client-004" | jq '.' || curl -s -X GET "$BASE_URL/api/clients/client-004"
echo ""
echo ""

echo "=========================================="
echo "Testing Complete"
echo "=========================================="


