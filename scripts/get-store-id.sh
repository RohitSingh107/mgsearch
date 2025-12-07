#!/bin/bash
# Quick script to get store ID from MongoDB database

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Load it from .env:"
    echo "  source .env  # or: export DATABASE_URL=..."
    exit 1
fi

echo "Store IDs in database:"
mongosh "$DATABASE_URL" --quiet --eval "
db.stores.find({}, {_id: 1, shop_domain: 1}).forEach(function(store) {
  print(store._id + ' | ' + store.shop_domain);
});
"

echo ""
echo "To generate a token, run:"
echo "  go run scripts/generate-token.go <store-id> <shop-domain>"
