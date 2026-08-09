#!/usr/bin/env bash

# Check if concurrently is installed
if ! command -v concurrently &> /dev/null; then
    echo "concurrently is not installed. Installing globally..."
    npm install -g concurrently
fi

echo "Starting all three GovMonitor services..."

concurrently -k -p "[{name}]" -n "BACKEND,FRONTEND,WA-SIDECAR" -c "blue,green,magenta" \
  "cd backend && go run ./cmd/api" \
  "cd frontend && npm run dev" \
  "cd backend/wa-sidecar && node index.js"
