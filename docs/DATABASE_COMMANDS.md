# Database Commands - Quick Reference (MongoDB)

## List All Collections

### Quick Script
```bash
mongosh $DATABASE_URL --eval "db.getCollectionNames()"
```

### Direct Command
```bash
mongosh $DATABASE_URL --eval "show collections"
```

---

## View Collection Structure

### Stores Collection
```bash
mongosh $DATABASE_URL --eval "db.stores.findOne()"
```

### Sessions Collection
```bash
mongosh $DATABASE_URL --eval "db.sessions.findOne()"
```

### All Collections with Document Counts
```bash
mongosh $DATABASE_URL --eval "
db.getCollectionNames().forEach(function(collection) {
    print(collection + ': ' + db[collection].countDocuments());
});
"
```

---

## View Collection Data

### All Stores
```bash
mongosh $DATABASE_URL --eval "db.stores.find({}, {shop_domain: 1, shop_name: 1, status: 1, installed_at: 1}).pretty()"
```

### All Sessions
```bash
mongosh $DATABASE_URL --eval "db.sessions.find({}, {id: 1, shop: 1, is_online: 1, expires: 1, created_at: 1}).pretty()"
```

### Find Store by Shop Domain
```bash
mongosh $DATABASE_URL --eval "db.stores.findOne({shop_domain: 'your-shop.myshopify.com'})"
```

### Find Sessions by Shop
```bash
mongosh $DATABASE_URL --eval "db.sessions.find({shop: 'your-shop.myshopify.com'}).pretty()"
```

---

## Create/Update Documents

### Create a Store (using mongosh)
```javascript
mongosh $DATABASE_URL
use mgsearch
db.stores.insertOne({
  shop_domain: "mg-store-207095.myshopify.com",
  shop_name: "Mg Store",
  encrypted_access_token: Buffer.from("dummy_token", "utf8"),
  api_key_public: "abc123def4567890abcdef1234567890",
  api_key_private: "private_key_here_32_chars_minimum_required",
  product_index_uid: "products_mg_store_207095_myshopify_com",
  meilisearch_index_uid: "products_mg_store_207095_myshopify_com",
  meilisearch_document_type: "product",
  meilisearch_url: "https://your-meilisearch-url.com",
  plan_level: "free",
  status: "active",
  webhook_secret: "webhook_secret_here_32_chars_minimum_required",
  sync_state: {status: "pending_initial_sync"},
  installed_at: new Date(),
  created_at: new Date(),
  updated_at: new Date()
})
```

### Update Store
```bash
mongosh $DATABASE_URL --eval "
db.stores.updateOne(
  {shop_domain: 'your-shop.myshopify.com'},
  {\$set: {status: 'active', updated_at: new Date()}}
)
"
```

---

## Delete Documents

### Delete a Store
```bash
mongosh $DATABASE_URL --eval "db.stores.deleteOne({shop_domain: 'your-shop.myshopify.com'})"
```

### Delete Expired Sessions
```bash
mongosh $DATABASE_URL --eval "db.sessions.deleteMany({expires: {\$lt: new Date()}})"
```

---

## Indexes

### List All Indexes
```bash
mongosh $DATABASE_URL --eval "db.stores.getIndexes()"
mongosh $DATABASE_URL --eval "db.sessions.getIndexes()"
```

### Create Index (if needed manually)
```bash
mongosh $DATABASE_URL --eval "db.stores.createIndex({shop_domain: 1}, {unique: true})"
mongosh $DATABASE_URL --eval "db.stores.createIndex({api_key_public: 1}, {unique: true})"
mongosh $DATABASE_URL --eval "db.sessions.createIndex({shop: 1})"
mongosh $DATABASE_URL --eval "db.sessions.createIndex({expires: 1})"
```

---

## Database Statistics

### Collection Stats
```bash
mongosh $DATABASE_URL --eval "db.stats()"
```

### Store Count
```bash
mongosh $DATABASE_URL --eval "db.stores.countDocuments()"
```

### Session Count
```bash
mongosh $DATABASE_URL --eval "db.sessions.countDocuments()"
```

---

## Backup and Restore

### Export Collection
```bash
mongoexport --uri="$DATABASE_URL" --collection=stores --out=stores.json
mongoexport --uri="$DATABASE_URL" --collection=sessions --out=sessions.json
```

### Import Collection
```bash
mongoimport --uri="$DATABASE_URL" --collection=stores --file=stores.json
mongoimport --uri="$DATABASE_URL" --collection=sessions --file=sessions.json
```

---

## Connection String Format

The `DATABASE_URL` should be in MongoDB format:
```
mongodb://localhost:27017/mgsearch
```

For authentication:
```
mongodb://username:password@localhost:27017/mgsearch
```
