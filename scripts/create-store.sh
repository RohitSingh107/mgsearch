#!/bin/bash
# Create a store with default values for testing (MongoDB)

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not set"
    echo "Load it from .env:"
    echo "  source .env"
    exit 1
fi

SHOP_DOMAIN="${1:-mg-store-207095.myshopify.com}"
SHOP_NAME="${2:-Mg Store}"

echo "Creating store: $SHOP_DOMAIN"

# Generate default values
PUBLIC_KEY=$(openssl rand -hex 16)
PRIVATE_KEY=$(openssl rand -hex 32)
WEBHOOK_SECRET=$(openssl rand -hex 32)
INDEX_UID="products_$(echo $SHOP_DOMAIN | sed 's/\./_/g')"

# Get encryption key from env (needed for dummy encrypted token)
if [ -z "$ENCRYPTION_KEY" ]; then
    echo "Warning: ENCRYPTION_KEY not set, using dummy encrypted token"
    ENCRYPTED_TOKEN_HEX=$(openssl rand -hex 32)
else
    # For a real token, you'd encrypt it, but for testing we'll use a dummy
    ENCRYPTED_TOKEN_HEX=$(openssl rand -hex 32)
fi

# Get Meilisearch URL from existing store or use default
MEILISEARCH_URL=$(mongosh "$DATABASE_URL" --quiet --eval "db.stores.findOne({}, {meilisearch_url: 1})?.meilisearch_url || 'https://your-meilisearch-url.com'")

# Create store document
mongosh "$DATABASE_URL" --quiet <<EOF
db.stores.updateOne(
  {shop_domain: "$SHOP_DOMAIN"},
  {
    \$set: {
      shop_name: "$SHOP_NAME",
      encrypted_access_token: BinData(0, "$ENCRYPTED_TOKEN_HEX"),
      api_key_public: "$PUBLIC_KEY",
      api_key_private: "$PRIVATE_KEY",
      product_index_uid: "$INDEX_UID",
      meilisearch_index_uid: "$INDEX_UID",
      meilisearch_document_type: "product",
      meilisearch_url: "$MEILISEARCH_URL",
      plan_level: "free",
      status: "active",
      webhook_secret: "$WEBHOOK_SECRET",
      sync_state: {status: "pending_initial_sync"},
      updated_at: new Date()
    },
    \$setOnInsert: {
      installed_at: new Date(),
      created_at: new Date()
    }
  },
  {upsert: true}
);
EOF

echo ""
echo "Store created! Here's your info:"
mongosh "$DATABASE_URL" --quiet --eval "
var store = db.stores.findOne({shop_domain: '$SHOP_DOMAIN'});
if (store) {
  print('ID: ' + store._id);
  print('Shop Domain: ' + store.shop_domain);
  print('API Key Public: ' + store.api_key_public);
} else {
  print('Store not found');
}
"
