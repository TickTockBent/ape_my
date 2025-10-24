#!/bin/bash
# End-to-end test for Ape_my v0.1.0
# This script tests the full CRUD workflow with a real running server

set -e

echo "=== Ape_my v0.1.0 End-to-End Test ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Build the binary
echo "Building ape_my..."
go build -o bin/ape_my ./cmd/ape_my
echo -e "${GREEN}✓${NC} Build successful"
echo

# Start the server in the background
echo "Starting server..."
./bin/ape_my examples/todos_schema.json with examples/todos_seed.json &
SERVER_PID=$!

# Wait for server to start
sleep 1

# Trap to ensure server is killed on exit
trap "kill $SERVER_PID 2>/dev/null || true" EXIT

echo -e "${GREEN}✓${NC} Server started (PID: $SERVER_PID)"
echo

# Test 1: List all todos (should have 3 from seed data)
echo "Test 1: GET /todos (list all)"
RESPONSE=$(curl -s http://localhost:8080/todos)
COUNT=$(echo $RESPONSE | grep -o '"id"' | wc -l)
if [ "$COUNT" -eq 3 ]; then
    echo -e "${GREEN}✓${NC} Returned 3 seeded todos"
else
    echo -e "${RED}✗${NC} Expected 3 todos, got $COUNT"
    exit 1
fi
echo

# Test 2: Get a specific todo
echo "Test 2: GET /todos/1 (get specific todo)"
RESPONSE=$(curl -s http://localhost:8080/todos/1)
if echo "$RESPONSE" | grep -q '"task"'; then
    echo -e "${GREEN}✓${NC} Successfully retrieved todo #1"
else
    echo -e "${RED}✗${NC} Failed to retrieve todo"
    exit 1
fi
echo

# Test 3: Create a new todo
echo "Test 3: POST /todos (create new todo)"
RESPONSE=$(curl -s -X POST http://localhost:8080/todos \
    -H "Content-Type: application/json" \
    -d '{"task":"Test task","completed":false,"priority":5}')
NEW_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -n "$NEW_ID" ]; then
    echo -e "${GREEN}✓${NC} Created new todo with ID: $NEW_ID"
else
    echo -e "${RED}✗${NC} Failed to create todo"
    exit 1
fi
echo

# Test 4: Update the new todo with PUT
echo "Test 4: PUT /todos/$NEW_ID (replace todo)"
RESPONSE=$(curl -s -X PUT http://localhost:8080/todos/$NEW_ID \
    -H "Content-Type: application/json" \
    -d '{"task":"Updated task","completed":true,"priority":1}')
if echo "$RESPONSE" | grep -q '"task":"Updated task"'; then
    echo -e "${GREEN}✓${NC} Successfully updated todo"
else
    echo -e "${RED}✗${NC} Failed to update todo"
    exit 1
fi
echo

# Test 5: Partially update with PATCH
echo "Test 5: PATCH /todos/$NEW_ID (partial update)"
RESPONSE=$(curl -s -X PATCH http://localhost:8080/todos/$NEW_ID \
    -H "Content-Type: application/json" \
    -d '{"priority":10}')
if echo "$RESPONSE" | grep -q '"priority":10'; then
    echo -e "${GREEN}✓${NC} Successfully patched todo"
else
    echo -e "${RED}✗${NC} Failed to patch todo"
    exit 1
fi
# Verify task is still "Updated task"
if echo "$RESPONSE" | grep -q '"task":"Updated task"'; then
    echo -e "${GREEN}✓${NC} Other fields preserved during patch"
else
    echo -e "${RED}✗${NC} Patch overwrote other fields"
    exit 1
fi
echo

# Test 6: Delete the todo
echo "Test 6: DELETE /todos/$NEW_ID"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE http://localhost:8080/todos/$NEW_ID)
if [ "$STATUS" -eq 204 ]; then
    echo -e "${GREEN}✓${NC} Successfully deleted todo (204 No Content)"
else
    echo -e "${RED}✗${NC} Expected 204, got $STATUS"
    exit 1
fi
echo

# Test 7: Verify deletion
echo "Test 7: GET /todos/$NEW_ID (verify deletion)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/todos/$NEW_ID)
if [ "$STATUS" -eq 404 ]; then
    echo -e "${GREEN}✓${NC} Todo not found (404) - deletion confirmed"
else
    echo -e "${RED}✗${NC} Expected 404, got $STATUS"
    exit 1
fi
echo

# Test 8: Verify list count is still 3 (original seeded data)
echo "Test 8: GET /todos (verify count after deletion)"
RESPONSE=$(curl -s http://localhost:8080/todos)
COUNT=$(echo $RESPONSE | grep -o '"id"' | wc -l)
if [ "$COUNT" -eq 3 ]; then
    echo -e "${GREEN}✓${NC} Todo count back to 3 (seeded data only)"
else
    echo -e "${RED}✗${NC} Expected 3 todos, got $COUNT"
    exit 1
fi
echo

# Test 9: Test 404 on unknown route
echo "Test 9: GET /unknown (test 404 handling)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/unknown)
if [ "$STATUS" -eq 404 ]; then
    echo -e "${GREEN}✓${NC} Returns 404 for unknown route"
else
    echo -e "${RED}✗${NC} Expected 404, got $STATUS"
    exit 1
fi
echo

# Test 10: Test content-type validation
echo "Test 10: POST without Content-Type (test 415)"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/todos -d '{"task":"test"}')
if [ "$STATUS" -eq 415 ]; then
    echo -e "${GREEN}✓${NC} Returns 415 Unsupported Media Type"
else
    echo -e "${RED}✗${NC} Expected 415, got $STATUS"
    exit 1
fi
echo

echo "================================================"
echo -e "${GREEN}ALL TESTS PASSED!${NC}"
echo "================================================"
echo
echo "Ape_my v0.1.0 is working perfectly!"
echo

# Cleanup happens automatically via trap
