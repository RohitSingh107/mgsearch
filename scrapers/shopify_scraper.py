#!/usr/bin/env python3
"""
Shopify sitemap scraper
Scrapes product URLs from Shopify sitemap.xml and saves product data to MongoDB
"""

import requests
import json
import sys
import os
from datetime import datetime
from typing import Dict, List, Any, Optional
from urllib.parse import urlparse, urljoin
from xml.etree import ElementTree as ET
from pymongo import MongoClient
from pymongo.errors import ConnectionFailure, DuplicateKeyError


class ShopifySitemapScraper:
    """Scraper for Shopify stores via sitemap.xml"""
    
    def __init__(self, mongo_uri: Optional[str] = None):
        """
        Initialize the scraper with MongoDB connection
        
        Args:
            mongo_uri: MongoDB connection URI (defaults to mongodb://localhost:27017/mgsearch)
        """
        self.mongo_uri = mongo_uri or os.getenv("DATABASE_URL", "mongodb://localhost:27017/mgsearch")
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36',
            'Accept': 'application/json, text/xml, */*',
            'Accept-Language': 'en-US,en;q=0.9',
        })
        
        # Connect to MongoDB
        try:
            self.mongo_client = MongoClient(self.mongo_uri)
            # Test connection
            self.mongo_client.admin.command('ping')
            print("✓ Connected to MongoDB")
        except ConnectionFailure as e:
            print(f"✗ Failed to connect to MongoDB: {e}")
            sys.exit(1)
    
    def extract_domain(self, url: str) -> str:
        """
        Extract domain name from URL for collection naming
        
        Args:
            url: Full URL
            
        Returns:
            Domain name (e.g., 'bananaclub.co.in')
        """
        parsed = urlparse(url)
        domain = parsed.netloc
        # Remove 'www.' if present
        if domain.startswith('www.'):
            domain = domain[4:]
        return domain
    
    def parse_sitemap_index(self, sitemap_url: str) -> List[str]:
        """
        Parse the main sitemap index to find product sitemap URLs
        
        Args:
            sitemap_url: URL to the main sitemap.xml
            
        Returns:
            List of product sitemap URLs
        """
        print(f"\n📋 Fetching sitemap index: {sitemap_url}")
        try:
            response = self.session.get(sitemap_url, timeout=30)
            response.raise_for_status()
            
            # Parse XML
            root = ET.fromstring(response.content)
            
            # Handle namespaces
            namespaces = {
                'sitemap': 'http://www.sitemaps.org/schemas/sitemap/0.9'
            }
            
            product_sitemaps = []
            
            # Find all sitemap entries
            for sitemap in root.findall('.//sitemap:sitemap', namespaces):
                loc = sitemap.find('sitemap:loc', namespaces)
                if loc is not None and loc.text:
                    sitemap_url = loc.text.strip()
                    # Only include product sitemaps
                    if 'sitemap_products' in sitemap_url:
                        product_sitemaps.append(sitemap_url)
                        print(f"  ✓ Found product sitemap: {sitemap_url}")
            
            print(f"📊 Found {len(product_sitemaps)} product sitemap(s)")
            return product_sitemaps
            
        except requests.RequestException as e:
            print(f"✗ Error fetching sitemap index: {e}")
            raise
        except ET.ParseError as e:
            print(f"✗ Error parsing XML: {e}")
            raise
    
    def parse_product_sitemap(self, sitemap_url: str) -> List[str]:
        """
        Parse a product sitemap to extract product URLs
        
        Args:
            sitemap_url: URL to a product sitemap
            
        Returns:
            List of product URLs
        """
        print(f"  📋 Parsing product sitemap: {sitemap_url}")
        try:
            response = self.session.get(sitemap_url, timeout=30)
            response.raise_for_status()
            
            # Parse XML
            root = ET.fromstring(response.content)
            
            # Handle namespaces
            namespaces = {
                'url': 'http://www.sitemaps.org/schemas/sitemap/0.9'
            }
            
            product_urls = []
            
            # Find all URL entries
            for url_elem in root.findall('.//url:url', namespaces):
                loc = url_elem.find('url:loc', namespaces)
                if loc is not None and loc.text:
                    product_url = loc.text.strip()
                    # Only include /products/ URLs
                    if '/products/' in product_url:
                        product_urls.append(product_url)
            
            print(f"    ✓ Found {len(product_urls)} product URL(s)")
            return product_urls
            
        except requests.RequestException as e:
            print(f"    ✗ Error fetching product sitemap: {e}")
            return []
        except ET.ParseError as e:
            print(f"    ✗ Error parsing XML: {e}")
            return []
    
    def fetch_product_json(self, product_url: str) -> Optional[Dict[str, Any]]:
        """
        Fetch product JSON data by appending .js to the product URL
        
        Args:
            product_url: Product page URL
            
        Returns:
            Product JSON data or None if error
        """
        # Append .js to the URL
        json_url = product_url.rstrip('/') + '.js'
        
        try:
            response = self.session.get(json_url, timeout=30)
            response.raise_for_status()
            
            # Parse JSON response
            product_data = response.json()
            return product_data
            
        except requests.RequestException as e:
            print(f"    ✗ Error fetching product JSON from {json_url}: {e}")
            return None
        except json.JSONDecodeError as e:
            print(f"    ✗ Error parsing JSON from {json_url}: {e}")
            return None
    
    def save_to_mongodb(self, products: List[Dict[str, Any]], collection_name: str):
        """
        Save products to MongoDB collection
        
        Args:
            products: List of product dictionaries
            collection_name: Name of the MongoDB collection
        """
        if not products:
            print("⚠ No products to save")
            return
        
        # Get database from connection URI
        db_name = self.mongo_uri.split('/')[-1].split('?')[0] if '/' in self.mongo_uri else 'mgsearch'
        db = self.mongo_client[db_name]
        collection = db[collection_name]
        
        # Create index on product ID for faster lookups
        try:
            collection.create_index("id", unique=True)
        except Exception:
            pass  # Index might already exist
        
        saved_count = 0
        updated_count = 0
        error_count = 0
        
        print(f"\n💾 Saving {len(products)} products to MongoDB collection: {collection_name}")
        
        for product in products:
            try:
                # Use product ID as the unique identifier
                product_id = product.get('id')
                if not product_id:
                    print(f"    ⚠ Skipping product without ID: {product.get('handle', 'unknown')}")
                    error_count += 1
                    continue
                
                # Add metadata
                product['_scraped_at'] = datetime.utcnow()
                
                # Upsert: insert if new, update if exists
                result = collection.update_one(
                    {'id': product_id},
                    {'$set': product},
                    upsert=True
                )
                
                if result.upserted_id:
                    saved_count += 1
                else:
                    updated_count += 1
                    
            except DuplicateKeyError:
                # This shouldn't happen with upsert, but handle it just in case
                updated_count += 1
            except Exception as e:
                print(f"    ✗ Error saving product {product.get('id', 'unknown')}: {e}")
                error_count += 1
        
        print(f"\n📊 Save Summary:")
        print(f"  ✓ New products: {saved_count}")
        print(f"  ✓ Updated products: {updated_count}")
        print(f"  ✗ Errors: {error_count}")
        print(f"  📦 Total in collection: {collection.count_documents({})}")
    
    def scrape(self, sitemap_url: str) -> List[Dict[str, Any]]:
        """
        Main scraping method: parse sitemap and fetch all products
        
        Args:
            sitemap_url: URL to the main sitemap.xml
            
        Returns:
            List of product dictionaries
        """
        print(f"\n{'='*80}")
        print(f"🚀 STARTING SHOPIFY SITEMAP SCRAPER")
        print(f"{'='*80}")
        print(f"📍 Sitemap URL: {sitemap_url}")
        
        # Extract domain for collection naming
        domain = self.extract_domain(sitemap_url)
        print(f"🌐 Domain: {domain}")
        print(f"📦 Collection: {domain}")
        
        # Step 1: Parse sitemap index to find product sitemaps
        product_sitemaps = self.parse_sitemap_index(sitemap_url)
        
        if not product_sitemaps:
            print("⚠ No product sitemaps found")
            return []
        
        # Step 2: Parse each product sitemap to get product URLs
        all_product_urls = []
        for sitemap_url in product_sitemaps:
            product_urls = self.parse_product_sitemap(sitemap_url)
            all_product_urls.extend(product_urls)
        
        print(f"\n📊 Total product URLs found: {len(all_product_urls)}")
        
        if not all_product_urls:
            print("⚠ No product URLs found")
            return []
        
        # Step 3: Fetch JSON data for each product
        print(f"\n🔄 Fetching product data...")
        products = []
        for idx, product_url in enumerate(all_product_urls, 1):
            print(f"  [{idx}/{len(all_product_urls)}] Fetching: {product_url}")
            product_data = self.fetch_product_json(product_url)
            if product_data:
                products.append(product_data)
        
        print(f"\n✓ Successfully fetched {len(products)}/{len(all_product_urls)} products")
        
        # Step 4: Save to MongoDB
        if products:
            self.save_to_mongodb(products, domain)
        
        print(f"\n{'='*80}")
        print(f"🎉 SCRAPING COMPLETE")
        print(f"{'='*80}")
        print(f"📊 Products scraped: {len(products)}")
        print(f"📦 MongoDB collection: {domain}")
        print(f"{'='*80}\n")
        
        return products


def main():
    """Main function to run the scraper"""
    if len(sys.argv) < 2:
        print("Usage: python shopify_scraper.py <sitemap_url>")
        print("Example: python shopify_scraper.py https://bananaclub.co.in/sitemap.xml")
        sys.exit(1)
    
    sitemap_url = sys.argv[1]
    
    # Validate URL
    if not sitemap_url.startswith('http'):
        print(f"✗ Invalid URL: {sitemap_url}")
        print("  Please provide a full URL starting with http:// or https://")
        sys.exit(1)
    
    # Get MongoDB URI from environment or use default
    mongo_uri = os.getenv("DATABASE_URL", "mongodb://localhost:27017/mgsearch")
    
    # Create scraper and run
    scraper = ShopifySitemapScraper(mongo_uri=mongo_uri)
    
    try:
        scraper.scrape(sitemap_url)
    except KeyboardInterrupt:
        print("\n\n⚠ Scraping interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n✗ Fatal error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
    finally:
        # Close MongoDB connection
        scraper.mongo_client.close()


if __name__ == "__main__":
    main()

