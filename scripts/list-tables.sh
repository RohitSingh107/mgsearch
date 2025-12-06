#!/bin/bash
# List all tables in the database

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Load it from .env:"
    echo "  source .env  # or: export DATABASE_URL=..."
    exit 1
fi

echo "=== All Tables in Database ==="
psql "$DATABASE_URL" -c "\dt"

echo ""
echo "=== Table Details ==="
echo ""
echo "--- Stores Table ---"
psql "$DATABASE_URL" -c "\d stores"

echo ""
echo "--- Sessions Table ---"
psql "$DATABASE_URL" -c "\d sessions"

echo ""
echo "=== Row Counts ==="
psql "$DATABASE_URL" -c "
SELECT 
    'stores' as table_name, 
    COUNT(*) as row_count 
FROM stores
UNION ALL
SELECT 
    'sessions' as table_name, 
    COUNT(*) as row_count 
FROM sessions;
"

