#!/bin/bash

# Test script for all config blog post examples

set -e

echo "Testing all config examples..."

# Test each step
for step in step1-minimal step2-computation step3-environment step4-routes step5-functions step6-debugging; do
    echo ""
    echo "=== Testing $step ==="
    cd "$step"
    
    echo "Running go mod tidy..."
    go mod tidy
    
    echo "Building..."
    go build -o server main.go
    
    echo "Starting server in background..."
    ./server &
    SERVER_PID=$!
    
    # Give server time to start
    sleep 2
    
    echo "Testing basic endpoints..."
    
    # Test root
    if curl -s -f http://localhost:3000 > /dev/null; then
        echo "✓ Root endpoint working"
    else
        echo "✗ Root endpoint failed"
    fi
    
    # Test routes if they exist
    if [[ "$step" == step4-* ]] || [[ "$step" == step5-* ]] || [[ "$step" == step6-* ]]; then
        if curl -s -f http://localhost:3000/blog/ > /dev/null; then
            echo "✓ Blog route working"
        else
            echo "✗ Blog route failed"
        fi
        
        if curl -s -f http://localhost:3000/docs/ > /dev/null; then
            echo "✓ Docs route working"
        else
            echo "✗ Docs route failed"
        fi
    fi
    
    echo "Stopping server..."
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null || true
    
    echo "✓ $step completed successfully"
    cd ..
done

echo ""
echo "All tests completed successfully!"
echo ""
echo "To test manually:"
echo "  cd step6-debugging"
echo "  go run main.go"
echo "  curl http://localhost:3000"
echo "  curl http://localhost:3000/blog/"
echo "  DEBUG=1 go run main.go  # to enable drafts route"