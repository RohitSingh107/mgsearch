-- Create a store with default values
-- Usage: psql $DATABASE_URL -f scripts/create-store-simple.sql

-- Set variables (modify these)
\set shop_domain 'mg-store-207095.myshopify.com'
\set shop_name 'Mg Store'

-- Generate keys (you can also set these manually)
\set public_key 'abc123def4567890abcdef1234567890'
\set private_key 'private_key_here_32_chars_minimum_required'
\set webhook_secret 'webhook_secret_here_32_chars_minimum_required'
\set index_uid 'products_mg_store_207095_myshopify_com'

-- Insert store
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
    :'shop_domain',
    :'shop_name',
    '\x' || encode(gen_random_bytes(32), 'hex'), -- Dummy encrypted token
    :'public_key',
    :'private_key',
    :'index_uid',
    :'index_uid',
    'product',
    (SELECT meilisearch_url FROM stores LIMIT 1), -- Use existing Meilisearch URL if available
    'free',
    'active',
    :'webhook_secret',
    '{"status": "pending_initial_sync"}'::jsonb
)
ON CONFLICT (shop_domain) DO UPDATE SET
    shop_name = EXCLUDED.shop_name,
    updated_at = NOW()
RETURNING id, shop_domain, api_key_public;

