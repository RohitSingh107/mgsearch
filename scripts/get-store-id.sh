#!/bin/bash
# Quick script to get store ID from database

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Load it from .env:"
    echo "  source .env  # or: export DATABASE_URL=..."
    exit 1
fi

echo "Store IDs in database:"
psql "$DATABASE_URL" -c "SELECT id, shop_domain FROM stores;"

echo ""
echo "To generate a token, run:"
echo "  go run scripts/generate-token.go <store-id> <shop-domain>"

