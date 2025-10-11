#!/bin/bash
# ACP Gateway Test Runner

set -e

echo "ACP Gateway Test Suite"
echo "======================"

# Check if servers are running
check_server() {
    local port=$1
    local name=$2
    
    if curl -s "http://localhost:$port/healthz" > /dev/null; then
        echo "✓ $name server is running on port $port"
        return 0
    else
        echo "✗ $name server is not running on port $port"
        return 1
    fi
}

# Test FastAPI
echo ""
echo "Testing FastAPI Implementation..."
if check_server 8080 "FastAPI"; then
    cd examples
    python test-fastapi.py
    echo "✓ FastAPI tests completed"
else
    echo "⚠ Skipping FastAPI tests - server not running"
    echo "  To start: cd fastapi && uvicorn main:app --reload --port 8080"
fi

# Test Express
echo ""
echo "Testing Express Implementation..."
if check_server 8080 "Express"; then
    cd examples
    if [ ! -f "node_modules/axios/package.json" ]; then
        echo "Installing axios..."
        npm install
    fi
    node test-express.js
    echo "✓ Express tests completed"
else
    echo "⚠ Skipping Express tests - server not running"
    echo "  To start: cd express && npm run dev"
fi

echo ""
echo "Test suite completed!"
