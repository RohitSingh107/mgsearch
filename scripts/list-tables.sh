#!/bin/bash
# List all collections in the MongoDB database

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Load it from .env:"
    echo "  source .env  # or: export DATABASE_URL=..."
    exit 1
fi

echo "=== All Collections in Database ==="
mongosh "$DATABASE_URL" --quiet --eval "db.getCollectionNames()"

echo ""
echo "=== Collection Details ==="
echo ""
echo "--- Stores Collection (sample document) ---"
mongosh "$DATABASE_URL" --quiet --eval "db.stores.findOne()"

echo ""
echo "--- Sessions Collection (sample document) ---"
mongosh "$DATABASE_URL" --quiet --eval "db.sessions.findOne()"

echo ""
echo "=== Document Counts ==="
mongosh "$DATABASE_URL" --quiet --eval "
print('stores: ' + db.stores.countDocuments());
print('sessions: ' + db.sessions.countDocuments());
"
