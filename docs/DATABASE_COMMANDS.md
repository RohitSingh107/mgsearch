# Database Commands - Quick Reference

## List All Tables

### Quick Script
```bash
./scripts/list-tables.sh
```

### Direct SQL
```bash
psql $DATABASE_URL -c "\dt"
```

### List with Details
```bash
psql $DATABASE_URL -c "\d+"
```

---

## View Table Structure

### Stores Table
```bash
psql $DATABASE_URL -c "\d stores"
```

### Sessions Table
```bash
psql $DATABASE_URL -c "\d sessions"
```

### All Tables with Details
```bash
psql $DATABASE_URL -c "\d+"
```

---

## View Table Data

### All Stores
```bash
psql $DATABASE_URL -c "SELECT id, shop_domain, shop_name, status, installed_at FROM stores;"
```

### All Sessions
```bash
psql $DATABASE_URL -c "SELECT id, shop, is_online, expires, created_at FROM sessions;"
```

### Stores with Keys (for getting storefront key)
```bash
psql $DATABASE_URL -c "SELECT id, shop_domain, api_key_public FROM stores;"
```

---

## Row Counts

### Count All Tables
```bash
psql $DATABASE_URL -c "
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
```

### Count Stores
```bash
psql $DATABASE_URL -c "SELECT COUNT(*) FROM stores;"
```

### Count Sessions
```bash
psql $DATABASE_URL -c "SELECT COUNT(*) FROM sessions;"
```

---

## Useful Queries

### Get Store by Domain
```bash
psql $DATABASE_URL -c "SELECT * FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';"
```

### Get Store ID (UUID)
```bash
psql $DATABASE_URL -t -c "SELECT id FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';"
```

### Get Storefront Key
```bash
psql $DATABASE_URL -t -c "SELECT api_key_public FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';"
```

### Get Sessions for a Shop
```bash
psql $DATABASE_URL -c "SELECT * FROM sessions WHERE shop = 'mg-store-207095.myshopify.com';"
```

---

## Connect to Database

### Using psql
```bash
psql $DATABASE_URL
```

### Interactive Session
```bash
psql $DATABASE_URL
# Then you can run SQL commands:
# \dt          - List tables
# \d stores    - Describe stores table
# SELECT * FROM stores;
# \q           - Quit
```

---

## Export Data

### Export Stores to CSV
```bash
psql $DATABASE_URL -c "COPY (SELECT * FROM stores) TO STDOUT WITH CSV HEADER" > stores.csv
```

### Export Sessions to CSV
```bash
psql $DATABASE_URL -c "COPY (SELECT * FROM sessions) TO STDOUT WITH CSV HEADER" > sessions.csv
```

---

## Common psql Commands

| Command | Description |
|---------|-------------|
| `\dt` | List all tables |
| `\d table_name` | Describe a table structure |
| `\d+ table_name` | Describe table with more details |
| `\l` | List all databases |
| `\c database_name` | Connect to a database |
| `\q` | Quit psql |
| `\?` | Show help |
| `\timing` | Toggle query timing |

---

## Quick Check Scripts

### Check if tables exist
```bash
psql $DATABASE_URL -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
```

### List all columns in stores table
```bash
psql $DATABASE_URL -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'stores';"
```

### Check indexes
```bash
psql $DATABASE_URL -c "\di"
```

---

## Troubleshooting

### If DATABASE_URL is not set:
```bash
# Load from .env
source .env

# Or set manually
export DATABASE_URL="postgres://mgsearch:mgsearch@localhost:5544/mgsearch?sslmode=disable"
```

### Test connection
```bash
psql $DATABASE_URL -c "SELECT version();"
```

### Check if database exists
```bash
psql $DATABASE_URL -c "SELECT current_database();"
```

