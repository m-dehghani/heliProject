#!/bin/bash
# be aware: if your OS is windows, run this file in "git bash" 
# run chmod +x test_integration_gateway_service.sh before run this script
# Base URL for the gateway service
BASE_URL="http://localhost:8080"

# Register a new user
echo "Registering a new user..."
curl -X POST "$BASE_URL/register" -H "Content-Type: application/json" -d '{
  "username": "testuser1",
  "password": "testpass1"
}'
echo -e "\n"

# Login the user
echo "Logging in the user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" -H "Content-Type: application/json" -d '{
  "username": "testuser1",
  "password": "testpass1"
}')
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"
echo -e "\n"

# Deposit money
echo "Depositing money..."
curl -X POST "$BASE_URL/deposit" -H "Content-Type: application/json" -H "Authorization: $TOKEN" -d '{
  "customer_id": 1,
  "amount": 100.0
}'
echo -e "\n"

# Withdraw money
echo "Withdrawing money..."
curl -X POST "$BASE_URL/withdraw" -H "Content-Type: application/json" -H "Authorization: $TOKEN" -d '{
  "customer_id": 1,
  "amount": 50.0
}'
echo -e "\n"

# Balance inquiry
echo "Inquiring balance..."
curl -X GET "$BASE_URL/balance?customer_id=1" -H "Authorization: $TOKEN"
echo -e "\n"

# Transaction history
echo "Fetching transaction history..."
curl -X GET "$BASE_URL/transactions?customer_id=1" -H "Authorization: $TOKEN"
echo -e "\n"

# Logout the user
echo "Logging out the user..."
curl -X POST "$BASE_URL/logout" -H "Authorization: $TOKEN"
echo -e "\n"
