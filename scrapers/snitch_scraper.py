#!/usr/bin/env python3
"""
Snitch.com scraper
Scrapes product search results from snitch.com and saves to CSV
"""

import requests
import csv
import json
import sys
from datetime import datetime
from typing import Dict, List, Any
from urllib.parse import quote


class SnitchScraper:
    """Scraper for snitch.com search API"""
    
    BASE_URL = "https://mxemjhp3rt.ap-south-1.awsapprunner.com"
    
    def __init__(self):
        """Initialize the scraper with required headers"""
        self.headers = {
            'Accept': 'application/json, text/plain, */*',
            'Accept-Headers': 'application/json',
            'Accept-Language': 'en-GB,en-US;q=0.9,en;q=0.8,hi;q=0.7,zu;q=0.6',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
            'DNT': '1',
            'Origin': 'https://www.snitch.com',
            'Pragma': 'no-cache',
            'Referer': 'https://www.snitch.com/',
            'Sec-Fetch-Dest': 'empty',
            'Sec-Fetch-Mode': 'cors',
            'Sec-Fetch-Site': 'cross-site',
            'User-Agent': 'Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36',
            'client-id': 'snitch_secret',
            'sec-ch-ua': '"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"',
            'sec-ch-ua-mobile': '?1',
            'sec-ch-ua-platform': '"Android"'
        }
        self.session = requests.Session()
        self.session.headers.update(self.headers)
    
    def search(self, query: str, page: int = 1, limit: int = 1000) -> Dict[str, Any]:
        """
        Search for products on snitch.com
        
        Args:
            query: Search query string
            page: Page number (default: 1)
            limit: Number of results per page (default: 100)
            
        Returns:
            JSON response from the API
        """
        # Build URL with query parameters
        params = {
            'page': page,
            'limit': limit,
            'keyword': query
        }
        url = f"{self.BASE_URL}/products/search"
        
        print(f"\n{'='*60}")
        print(f"REQUEST URL: {url}")
        print(f"QUERY PARAMS: {params}")
        print(f"{'='*60}")
        
        try:
            response = self.session.get(url, params=params, timeout=30)
            
            # Log response details
            print(f"\nRESPONSE STATUS CODE: {response.status_code}")
            print(f"CONTENT TYPE: {response.headers.get('Content-Type', 'Unknown')}")
            print(f"CONTENT LENGTH: {len(response.content)} bytes")
            
            response.raise_for_status()
            
            # Parse JSON response
            try:
                data = response.json()
                print(f"\n✓ Successfully parsed as JSON")
                
                # Log response structure summary
                if isinstance(data, dict):
                    print(f"Response keys: {list(data.keys())}")
                    if 'data' in data:
                        if isinstance(data['data'], dict):
                            print(f"'data' is a dict with keys: {list(data['data'].keys())}")
                            # Log pagination info if available
                            if 'total_count' in data['data']:
                                print(f"Total products available: {data['data']['total_count']}")
                            if 'page' in data['data']:
                                print(f"Current page: {data['data']['page']}")
                            if 'limit' in data['data']:
                                print(f"Limit per page: {data['data']['limit']}")
                            if 'products' in data['data'] and isinstance(data['data']['products'], list):
                                print(f"Products in this page: {len(data['data']['products'])}")
                        elif isinstance(data['data'], list):
                            print(f"Found {len(data['data'])} items in 'data' array")
                    if 'products' in data:
                        if isinstance(data['products'], list):
                            print(f"Found {len(data['products'])} items in 'products' array")
                        else:
                            print(f"'products' is {type(data['products'])}")
                
                return data
            except json.JSONDecodeError as json_err:
                print(f"\n✗ JSON parsing error: {json_err}")
                print(f"Error position: {json_err.pos if hasattr(json_err, 'pos') else 'unknown'}")
                
                # Log raw response for debugging
                raw_text = response.text
                print(f"\nRAW RESPONSE (first 2000 chars):")
                print(f"{'='*60}")
                print(raw_text[:2000])
                print(f"{'='*60}")
                
                return {}
            
        except requests.exceptions.RequestException as e:
            print(f"\n✗ Request error: {e}", file=sys.stderr)
            if hasattr(e, 'response') and e.response is not None:
                print(f"Response status: {e.response.status_code}")
                print(f"Response text (first 500 chars): {e.response.text[:500]}")
            return {}
    
    def extract_products(self, data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """
        Extract product information from API response
        
        Args:
            data: JSON response from API
            
        Returns:
            List of product dictionaries
        """
        products = []
        
        print(f"\n{'='*60}")
        print("EXTRACTING PRODUCTS FROM RESPONSE...")
        print(f"{'='*60}")
        
        # The response structure may vary, so we'll try multiple possible paths
        if isinstance(data, dict):
            # Try to find products in various possible locations
            # Common API response structures: {data: [...], products: [...], results: [...]}
            if 'data' in data and isinstance(data['data'], list):
                print("Found 'data' key with list")
                products_data = data['data']
            elif 'products' in data and isinstance(data['products'], list):
                print("Found 'products' key with list")
                products_data = data['products']
            elif 'results' in data and isinstance(data['results'], list):
                print("Found 'results' key with list")
                products_data = data['results']
            elif 'items' in data and isinstance(data['items'], list):
                print("Found 'items' key with list")
                products_data = data['items']
            elif 'data' in data and isinstance(data['data'], dict):
                # Sometimes data is nested: {data: {products: [...]}}
                nested_data = data['data']
                if 'products' in nested_data:
                    print("Found nested 'data.products'")
                    products_data = nested_data['products']
                elif 'results' in nested_data:
                    print("Found nested 'data.results'")
                    products_data = nested_data['results']
                else:
                    print("Found 'data' key but it's a dict, searching recursively...")
                    products_data = self._find_products_recursive(nested_data)
            else:
                # If it's a nested structure, try to find arrays
                print("No direct product keys found, searching recursively...")
                products_data = self._find_products_recursive(data)
            
            print(f"Products data type: {type(products_data)}")
            
            if isinstance(products_data, list):
                print(f"Found list with {len(products_data)} items")
                for idx, item in enumerate(products_data):
                    print(f"  Processing item {idx+1}...")
                    product = self._extract_product_info(item)
                    if product:
                        print(f"    ✓ Extracted product: {product.get('title', product.get('id', 'unknown'))}")
                        products.append(product)
                    else:
                        print(f"    ✗ No product data extracted")
            elif isinstance(products_data, dict):
                print(f"Found dict with keys: {list(products_data.keys())[:10]}")
                # Sometimes products are nested in a dict
                for key, value in products_data.items():
                    if isinstance(value, list):
                        print(f"  Found list in key '{key}' with {len(value)} items")
                        for item in value:
                            product = self._extract_product_info(item)
                            if product:
                                products.append(product)
        
        # If still no products, try deeper recursive search for product-like objects
        if not products:
            print("\nNo products found with standard extraction, trying deep search...")
            all_objects = self._deep_search_for_products(data)
            print(f"Found {len(all_objects)} potential product objects")
            for obj in all_objects:
                product = self._extract_product_info(obj)
                if product and self._looks_like_product(product):
                    products.append(product)
        
        print(f"\nTotal products extracted: {len(products)}")
        print(f"{'='*60}\n")
        
        return products
    
    def _find_products_recursive(self, data: Any, depth: int = 0) -> List[Any]:
        """Recursively search for product arrays in nested structures"""
        if depth > 10:  # Increased depth limit for RSC structures
            return []
        
        if isinstance(data, list):
            # Check if this list contains product-like objects
            if len(data) > 0 and isinstance(data[0], dict):
                # Check if items look like products (have product-like fields)
                sample = data[0]
                product_indicators = ['title', 'name', 'price', 'id', 'product_id', 
                                    'image', 'url', 'productUrl', 'selling_price']
                if any(key in sample for key in product_indicators):
                    return data
        elif isinstance(data, dict):
            # Skip RSC metadata fields
            skip_keys = ['children', 'className', 'routeParams', '__PAGE__', 
                        'needHamBurger', 'needSearchBar', 'needDesktopIcons', 
                        'displayAppNudge', 'key', 'value']
            
            for key, value in data.items():
                if key in skip_keys:
                    continue
                    
                if isinstance(value, list) and len(value) > 0:
                    if isinstance(value[0], dict):
                        # Check if items look like products
                        sample = value[0]
                        product_indicators = ['title', 'name', 'price', 'id', 'product_id',
                                            'image', 'url', 'productUrl', 'selling_price']
                        if any(k in sample for k in product_indicators):
                            return value
                result = self._find_products_recursive(value, depth + 1)
                if result:
                    return result
        
        return []
    
    def _deep_search_for_products(self, data: Any, depth: int = 0, found: List[Any] = None) -> List[Any]:
        """Deep recursive search for any product-like objects"""
        if found is None:
            found = []
        
        if depth > 15:  # Deep limit
            return found
        
        if isinstance(data, dict):
            # Check if this dict looks like a product
            product_indicators = ['title', 'name', 'price', 'id', 'product_id',
                                'image', 'url', 'productUrl', 'selling_price', 
                                'product_name', 'productName']
            if any(key in data for key in product_indicators):
                # Make sure it's not just RSC metadata
                if 'children' not in data or 'className' not in data:
                    found.append(data)
            
            # Continue searching
            for key, value in data.items():
                if key in ['children', 'className', 'routeParams']:
                    continue
                self._deep_search_for_products(value, depth + 1, found)
        
        elif isinstance(data, list):
            for item in data:
                self._deep_search_for_products(item, depth + 1, found)
        
        return found
    
    def _looks_like_product(self, product: Dict[str, Any]) -> bool:
        """Check if extracted data looks like a product"""
        if not product:
            return False
        
        # Must have at least one of these key fields
        key_fields = ['title', 'name', 'id', 'product_id', 'price', 'selling_price']
        has_key_field = any(key in product for key in key_fields)
        
        # Should not be RSC metadata
        is_not_rsc = 'children' not in product and 'className' not in product
        
        return has_key_field and is_not_rsc
    
    def _extract_product_info(self, item: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extract ALL product information from a product item, including nested structures
        
        Args:
            item: Product item from API response
            
        Returns:
            Dictionary with all extracted product information (flattened)
        """
        product = {}
        
        # Process all fields from the item
        for key, value in item.items():
            if value is None:
                product[key] = None
            elif isinstance(value, bool):
                product[key] = value
            elif isinstance(value, (int, float)):
                product[key] = value
            elif isinstance(value, str):
                product[key] = value
            elif isinstance(value, dict):
                # Flatten nested dictionaries by converting to JSON string
                # This preserves all nested data like variants, location_info, etc.
                product[key] = json.dumps(value, ensure_ascii=False)
            elif isinstance(value, list):
                # Handle lists - if they contain simple types, join them
                # If they contain complex objects, convert to JSON
                if len(value) == 0:
                    product[key] = '[]'
                elif all(isinstance(v, (str, int, float, bool)) or v is None for v in value):
                    # Simple list of primitives
                    product[key] = ', '.join(str(v) if v is not None else '' for v in value)
                else:
                    # Complex list (e.g., variants array with objects)
                    product[key] = json.dumps(value, ensure_ascii=False)
            else:
                # Fallback: convert to string
                product[key] = str(value)
        
        return product
    
    def remove_duplicates_by_title(self, products: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Remove duplicate products based on title (case-insensitive)
        
        Args:
            products: List of product dictionaries
            
        Returns:
            List of unique products (keeps first occurrence)
        """
        seen_titles = set()
        unique_products = []
        duplicates_count = 0
        
        for product in products:
            # Get title and normalize it (lowercase, strip whitespace)
            title = product.get('title', '')
            if title:
                normalized_title = str(title).lower().strip()
            else:
                # If no title, use a combination of other fields or empty string
                normalized_title = ''
            
            # Skip if we've seen this title before
            if normalized_title and normalized_title in seen_titles:
                duplicates_count += 1
                continue
            
            # Add to seen set and keep product
            if normalized_title:
                seen_titles.add(normalized_title)
            unique_products.append(product)
        
        if duplicates_count > 0:
            print(f"✓ Removed {duplicates_count} duplicate products (based on title)")
        
        return unique_products
    
    def save_to_csv(self, products: List[Dict[str, Any]], filename: str = None) -> str:
        """
        Save products to CSV file
        
        Args:
            products: List of product dictionaries
            filename: Output CSV filename (optional)
            
        Returns:
            Path to the saved CSV file
        """
        if not filename:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"snitch_products_{timestamp}.csv"
        
        if not products:
            print("No products to save", file=sys.stderr)
            return filename
        
        # Remove duplicates before saving
        print(f"\n📊 Removing duplicates...")
        print(f"   Original count: {len(products)}")
        products = self.remove_duplicates_by_title(products)
        print(f"   After deduplication: {len(products)}")
        
        # Get all unique keys from all products
        all_keys = set()
        for product in products:
            all_keys.update(product.keys())
        
        # Sort keys for consistent column order
        fieldnames = sorted(all_keys)
        
        # Ensure common fields are first
        preferred_order = ['id', 'title', 'price', 'original_price', 'discount', 
                          'brand', 'category', 'url', 'image', 'description',
                          'in_stock', 'rating', 'reviews', 'sku', 'color', 
                          'size', 'material']
        
        # Reorder fieldnames
        ordered_fieldnames = []
        for field in preferred_order:
            if field in fieldnames:
                ordered_fieldnames.append(field)
                fieldnames.remove(field)
        ordered_fieldnames.extend(sorted(fieldnames))
        
        with open(filename, 'w', newline='', encoding='utf-8') as csvfile:
            writer = csv.DictWriter(csvfile, fieldnames=ordered_fieldnames, 
                                   extrasaction='ignore')
            writer.writeheader()
            
            for product in products:
                # Convert all values to strings and handle None values
                row = {}
                for key in ordered_fieldnames:
                    value = product.get(key, '')
                    if value is None:
                        row[key] = ''
                    elif isinstance(value, (dict, list)):
                        row[key] = json.dumps(value)
                    else:
                        row[key] = str(value)
                writer.writerow(row)
        
        print(f"Saved {len(products)} products to {filename}")
        return filename


def scrape_query(scraper: SnitchScraper, query: str, limit: int = 1000) -> List[Dict[str, Any]]:
    """
    Scrape all products for a single query with pagination
    
    Args:
        scraper: SnitchScraper instance
        query: Search query string
        limit: Number of results per page
        
    Returns:
        List of all product dictionaries for this query
    """
    print(f"\n{'='*80}")
    print(f"🔍 SEARCHING FOR: '{query}'")
    print(f"{'='*80}")
    
    page = 1
    all_products = []
    total_count = None
    total_pages = None
    
    print(f"\n{'='*60}")
    print(f"STARTING PAGINATION - Fetching ALL products for '{query}'")
    print(f"{'='*60}\n")
    
    while True:
        print(f"\n{'='*60}")
        print(f"Query: '{query}' | Page {page} (limit: {limit})")
        if total_count is not None:
            print(f"Progress: {len(all_products)}/{total_count} products collected")
        if total_pages is not None:
            print(f"Page {page} of {total_pages}")
        print(f"{'='*60}")
        
        response_data = scraper.search(query, page=page, limit=limit)
        
        # Extract products from this page
        products = scraper.extract_products(response_data)
        
        if not products:
            print(f"No products found on page {page}")
            break
        
        all_products.extend(products)
        print(f"\n✓ Collected {len(products)} products from page {page}")
        print(f"✓ Total products collected for '{query}': {len(all_products)}")
        
        # Check pagination info from API response
        # The API structure is: {data: {products: [...], total_count: X, page: Y, limit: Z}}
        has_more = False
        if isinstance(response_data, dict) and 'data' in response_data:
            data_obj = response_data['data']
            if isinstance(data_obj, dict):
                # Get total_count if available
                if 'total_count' in data_obj:
                    total_count = data_obj['total_count']
                    # Calculate total pages
                    if total_count and limit:
                        total_pages = (total_count + limit - 1) // limit  # Ceiling division
                        print(f"📊 Total products available: {total_count}")
                        print(f"📊 Total pages: {total_pages}")
                
                # Check if there are more pages
                if total_count is not None:
                    # Use total_count to determine if more pages exist
                    if len(all_products) < total_count:
                        has_more = True
                    else:
                        has_more = False
                        print(f"✓ Reached total count: {len(all_products)}/{total_count}")
                elif 'hasMore' in data_obj:
                    has_more = data_obj['hasMore']
                elif 'nextPage' in data_obj and data_obj['nextPage']:
                    has_more = True
                elif len(products) == limit:
                    # If we got a full page, assume there might be more
                    has_more = True
                    print(f"⚠ No pagination info found, assuming more pages (got full page of {limit})")
        
        # Also check top-level pagination fields
        if not has_more and isinstance(response_data, dict):
            if 'hasMore' in response_data and response_data['hasMore']:
                has_more = True
            elif 'totalPages' in response_data:
                total_pages = response_data['totalPages']
                if page < total_pages:
                    has_more = True
            elif 'nextPage' in response_data and response_data['nextPage']:
                has_more = True
            elif len(products) == limit and total_count is None:
                # Fallback: if we got a full page and don't know total, assume more
                has_more = True
        
        if not has_more:
            print(f"\n{'='*60}")
            print(f"✓ Pagination complete for '{query}' - No more pages to fetch")
            print(f"{'='*60}")
            break
        
        page += 1
        
        # Safety limit to prevent infinite loops
        if page > 1000:
            print(f"\n⚠ Safety limit reached: Stopped at page {page}")
            break
    
    print(f"\n{'='*80}")
    print(f"✓ COMPLETED: '{query}' - Found {len(all_products)} products")
    print(f"{'='*80}\n")
    
    return all_products


def main():
    """Main function to run the scraper with multiple queries"""
    scraper = SnitchScraper()
    
    # Define array of search queries
    # You can modify this array or pass queries via command line
    search_queries = [
        "sneakers",
        "t shirt",
        "jeans",
        "hoodie",
        "jacket"
    ]
    
    # Allow queries to be passed as command line arguments
    # If arguments provided, use them instead of default array
    if len(sys.argv) > 1:
        # Command line format: python scraper.py "query1" "query2" "query3"
        search_queries = sys.argv[1:]
    
    print(f"\n{'='*80}")
    print(f"🚀 STARTING SCRAPER")
    print(f"{'='*80}")
    print(f"📋 Total queries to process: {len(search_queries)}")
    queries_str = ', '.join(f"'{q}'" for q in search_queries)
    print(f"📋 Queries: {queries_str}")
    print(f"{'='*80}\n")
    
    limit = 1000
    all_products = []
    
    # Process each query
    for idx, query in enumerate(search_queries, 1):
        print(f"\n{'#'*80}")
        print(f"# Processing query {idx}/{len(search_queries)}")
        print(f"{'#'*80}")
        
        try:
            products = scrape_query(scraper, query, limit)
            all_products.extend(products)
            print(f"✓ Query '{query}' completed: {len(products)} products")
            print(f"✓ Total products collected so far: {len(all_products)}")
        except Exception as e:
            print(f"\n✗ Error processing query '{query}': {e}", file=sys.stderr)
            continue
    
    if not all_products:
        print("\n⚠ No products found across all queries.")
        return
    
    print(f"\n{'='*80}")
    print(f"🎉 SCRAPING COMPLETE")
    print(f"{'='*80}")
    print(f"📊 Total queries processed: {len(search_queries)}")
    print(f"📊 Total products found: {len(all_products)}")
    print(f"{'='*80}\n")
    
    # Save to CSV
    csv_filename = scraper.save_to_csv(all_products)
    print(f"💾 Data saved to: {csv_filename}")


if __name__ == "__main__":
    main()

