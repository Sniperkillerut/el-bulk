#!/bin/bash
echo "Starting stress tests..."
echo "1. Client load (GET /api/products)"
for i in {1..50}; do
  curl -s "http://localhost:8080/api/products?page=1&limit=50" > /dev/null &
done
wait
echo "Client load done."

echo "2. Admin Sync Load (POST /api/admin/prices/refresh)"
# Assuming we need an admin token, but let's see if it's protected in dev.
for i in {1..10}; do
  curl -s -X POST "http://localhost:8080/api/admin/prices/refresh" \
    -H "Authorization: Bearer dev-admin-token" > /dev/null &
done
wait
echo "Admin Sync done."

echo "3. Capturing docker stats..."
docker stats --no-stream
