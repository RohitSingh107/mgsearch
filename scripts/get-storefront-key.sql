-- Get Storefront API Key for a specific store
-- Usage: psql $DATABASE_URL -f scripts/get-storefront-key.sql

-- Option 1: Get key for specific store
SELECT 
    shop_domain,
    api_key_public AS storefront_key,
    shop_name,
    status,
    installed_at
FROM stores 
WHERE shop_domain = 'mg-store-207095.myshopify.com';

-- Option 2: List all stores with their keys
-- SELECT 
--     shop_domain,
--     api_key_public AS storefront_key,
--     shop_name,
--     status
-- FROM stores
-- ORDER BY installed_at DESC;

