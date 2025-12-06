#!/bin/bash
# Create a store with default values for testing

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
    ENCRYPTED_TOKEN="\\x$(openssl rand -hex 32)"
else
    # For a real token, you'd encrypt it, but for testing we'll use a dummy
    ENCRYPTED_TOKEN="\\x$(openssl rand -hex 32)"
fi

# Insert store with default values
psql "$DATABASE_URL" <<EOF
INSERT INTO stores (
    shop_domain,
    shop_name,
    encrypted_access_token,
    api_key_public,
    api_key_private,
    product_index_uid,
    meilisearch_index_uid,
    meilisearch_document_type,
    meilisearch_url,
    plan_level,
    status,
    webhook_secret,
    sync_state
) VALUES (
    '$SHOP_DOMAIN',
    '$SHOP_NAME',
    '$ENCRYPTED_TOKEN',
    '$PUBLIC_KEY',
    '$PRIVATE_KEY',
    '$INDEX_UID',
    '$INDEX_UID',
    'product',
    (SELECT meilisearch_url FROM stores LIMIT 1),
    'free',
    'active',
    '$WEBHOOK_SECRET',
    '{"status": "pending_initial_sync"}'::jsonb
)
ON CONFLICT (shop_domain) DO UPDATE SET
    shop_name = EXCLUDED.shop_name,
    updated_at = NOW()
RETURNING id, shop_domain, api_key_public;
EOF

echo ""
echo "Store created! Here's your info:"
psql "$DATABASE_URL" -c "SELECT id, shop_domain, api_key_public FROM stores WHERE shop_domain = '$SHOP_DOMAIN';"

