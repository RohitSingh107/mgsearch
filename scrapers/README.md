# Product Scrapers

Python scrapers for extracting product data from various e-commerce websites.

## Available Scrapers

- **Snitch.com Scraper** (`snitch_scraper.py`) - Scrapes product data from snitch.com
- **Tasva.com Scraper** (`tasva_scraper.py`) - Scrapes product data from Tasva.com
- **Shopify Sitemap Scraper** (`shopify_scraper.py`) - Scrapes product data from Shopify stores via sitemap.xml

## Setup

1. Install Python dependencies:
```bash
pip install -r requirements.txt
```

2. Set up MongoDB connection (optional, defaults to `mongodb://localhost:27017/mgsearch`):
```bash
export DATABASE_URL="mongodb://localhost:27017/mgsearch"
```

## Usage

### Snitch.com Scraper

Run the scraper with a search query:

```bash
python snitch_scraper.py "t shirt in red"
```

If no query is provided, it defaults to several predefined queries.

### Tasva.com Scraper

Run the scraper with a search query:

```bash
python tasva_scraper.py "kurta"
```

You can pass multiple queries:
```bash
python tasva_scraper.py "kurta" "shirt" "pants"
```

If no query is provided, it defaults to several predefined queries.

### Shopify Sitemap Scraper

Scrapes all products from a Shopify store by parsing the sitemap.xml file.

**Usage:**
```bash
python shopify_scraper.py <sitemap_url>
```

**Example:**
```bash
python shopify_scraper.py https://bananaclub.co.in/sitemap.xml
```

**How it works:**
1. Fetches the main sitemap.xml URL
2. Parses the sitemap index to find all product sitemap URLs (e.g., `sitemap_products_1.xml`, `sitemap_products_2.xml`)
3. Extracts all product URLs from each product sitemap
4. For each product URL, appends `.js` to fetch JSON data (e.g., `https://bananaclub.co.in/products/product-name.js`)
5. Saves all products to MongoDB in a collection named after the domain (e.g., `bananaclub.co.in`)

**MongoDB Collection:**
- Collection name is automatically set to the domain name (e.g., `bananaclub.co.in`)
- Products are stored with their full JSON data
- Duplicate products (by ID) are automatically updated
- Each product includes a `_scraped_at` timestamp

**Environment Variables:**
- `DATABASE_URL`: MongoDB connection string (default: `mongodb://localhost:27017/mgsearch`)

## Output

- **Snitch scraper** saves data to: `snitch_products_YYYYMMDD_HHMMSS.csv`
- **Tasva scraper** saves data to: `tasva_products_YYYYMMDD_HHMMSS.csv`
- **Shopify scraper** saves data to MongoDB collection named after the domain

CSV scrapers save files in the same directory as the script.

## Features

### Common Features (All Scrapers)
- Uses exact headers from API requests
- Extracts comprehensive product information (title, price, images, URLs, variants, etc.)
- Handles nested JSON structures
- Automatic pagination to fetch all products
- Duplicate removal
- Error handling and debugging output

### Snitch Scraper Specific
- GET-based API with query parameters
- Supports pagination via page numbers
- Saves data to CSV format

### Tasva Scraper Specific
- POST-based API with JSON payload
- Supports pagination via offset/limit
- Extracts detailed product features (fabric, fit, occasion, etc.)
- Extracts size variants and color variants
- Includes discount information
- Saves data to CSV format

### Shopify Scraper Specific
- Parses XML sitemaps to discover all products
- Automatically handles multiple product sitemap files
- Fetches product JSON by appending `.js` to product URLs
- Saves to MongoDB with domain-based collection naming
- Handles duplicate products (upsert by product ID)
- Includes scraping timestamp metadata
