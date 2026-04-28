#!/bin/bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzczNDYwNzgsInN1YiI6ImFkbWluLWlkIn0.Mu_aRFjzANhjUm-ttUpuM8O1Xz9Na9OnORoj4OYueek"
BASE_URL="http://localhost:8080"

echo "Starting Enhanced Stress Tests..."

echo "1. Client load: GET /api/products (50 concurrent requests)"
for i in {1..50}; do
  curl -s "${BASE_URL}/api/products?page=1&limit=50" > /dev/null &
done
wait
echo "Client load done."

echo "2. Admin Stats load: GET /api/admin/stats (20 concurrent requests)"
for i in {1..20}; do
  curl -s -H "Authorization: Bearer ${TOKEN}" -H "X-Requested-With: XMLHttpRequest" "${BASE_URL}/api/admin/stats" > /dev/null &
done
wait
echo "Admin Stats done."

echo "3. Bulk Price Update: POST /api/admin/prices/refresh (5 concurrent requests)"
for i in {1..5}; do
  curl -s -X POST -H "Authorization: Bearer ${TOKEN}" -H "X-Requested-With: XMLHttpRequest" "${BASE_URL}/api/admin/prices/refresh" > /dev/null &
done
wait
echo "Bulk Price Update done."

echo "4. TCG Sync: POST /api/admin/tcgs/mtg/sync-prices (2 concurrent requests)"
for i in {1..2}; do
  curl -s -X POST -H "Authorization: Bearer ${TOKEN}" -H "X-Requested-With: XMLHttpRequest" "${BASE_URL}/api/admin/tcgs/mtg/sync-prices" > /dev/null &
done
wait
echo "TCG Sync done."

echo "Final docker stats:"
docker stats el_bulk_backend_dev el_bulk_frontend_dev el_bulk_db --no-stream
