#!/usr/bin/env python3
"""
Tasva.com scraper
Scrapes product search results from Tasva API and saves to CSV
"""

import requests
import csv
import json
import sys
import hashlib
import time
from datetime import datetime
from typing import Dict, List, Any, Optional


class TasvaScraper:
    """Scraper for Tasva.com product API"""
    
    BASE_URL = "https://plpengineapis.abfrl.in/fetchProducts"
    
    def __init__(self, device_token: Optional[str] = None, secure_key: Optional[str] = None,
                 device_id: Optional[str] = None):
        """
        Initialize the scraper with required headers
        
        Args:
            device_token: Device token for authentication (if None, uses default)
            secure_key: Secure key for API (if None, uses default)
            device_id: Device ID (if None, uses default from curl command)
        """
        # Default values from the curl command (may need to be updated)
        self.device_token = device_token or "c7c647b4180847612266af3a9190ae4e72fe6ab36ab44cf13a936a7cc10b5cf9.1768676366"
        self.secure_key = secure_key or "00175e0566c5448c95b0d5e9051a6a00"
        # Use exact deviceId from curl command
        self.device_id = device_id or "1603584e4a91d840b878021fedc187f5"
        
        # Generate authorization token (JWT format - simplified for now)
        # In production, you might need to generate this properly
        self.auth_token = self._generate_auth_token()
        
        self.headers = {
            'Accept': 'application/json, text/plain, */*',
            'Accept-Language': 'en-GB,en-US;q=0.9,en;q=0.8,hi;q=0.7,zu;q=0.6',
            'Authorization': f'Bearer {self.auth_token}',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
            'Content-Type': 'application/json',
            'DNT': '1',
            'Origin': 'https://www.tasva.com',
            'Pragma': 'no-cache',
            'Referer': 'https://www.tasva.com/',
            'Sec-Fetch-Dest': 'empty',
            'Sec-Fetch-Mode': 'cors',
            'Sec-Fetch-Site': 'cross-site',
            'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36',
            'env': 'development',
            'requestBrand': 'TASVA',
            'sec-ch-ua': '"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"',
            'sec-ch-ua-mobile': '?0',
            'sec-ch-ua-platform': '"Linux"',
            'securekey': self.secure_key
        }
        self.session = requests.Session()
        self.session.headers.update(self.headers)
    
    def _generate_auth_token(self) -> str:
        """
        Generate authorization token
        For now, using the token from the curl command
        In production, you might need to implement proper token generation
        """
        # This is a placeholder - you may need to implement proper JWT generation
        # or fetch a fresh token from an auth endpoint
        return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlblBheWxvYWQiOnsiZGV2aWNlVG9rZW4iOiJjN2M2NDdiNDE4MDg0NzYxMjI2NmFmM2E5MTkwYWU0ZTcyZmU2YWIzNmFiNDRjZjEzYTkzNmE3Y2MxMGI1Y2Y5LjE3Njg2NzYzNjYifSwiaWF0IjoxNzY4Njc2MzY3fQ.tl2_7I8lkU6u_Jm6XPHiD6yK46oL51Vfup8MwL6B1Jk"
    
    def _generate_hash(self, payload: Dict[str, Any]) -> str:
        """
        Generate hash for request validation
        Based on reverse engineering from the curl command
        
        The hash from curl for "kurta" query: 30c23fb01d501a7bf6d4e68019a9992d
        This method tries multiple algorithms to find the correct one.
        
        Args:
            payload: Request payload dictionary
            
        Returns:
            MD5 hash string
        """
        # Create a copy without hash and validateHash for hashing
        payload_for_hash = {k: v for k, v in payload.items() 
                          if k not in ['hash', 'validateHash']}
        
        # Method 1: Hash of compact JSON string (sorted keys) + securekey
        payload_json = json.dumps(payload_for_hash, sort_keys=True, separators=(',', ':'))
        hash_string1 = payload_json + self.secure_key
        hash1 = hashlib.md5(hash_string1.encode()).hexdigest()
        
        # Method 2: Hash of compact JSON string + securekey (different order)
        hash_string2 = self.secure_key + payload_json
        hash2 = hashlib.md5(hash_string2.encode()).hexdigest()
        
        # Method 3: Key-value pairs in sorted order, no JSON formatting
        sorted_keys = sorted(payload_for_hash.keys())
        hash_parts = []
        for k in sorted_keys:
            v = payload_for_hash[k]
            if isinstance(v, bool):
                hash_parts.append(f"{k}:{str(v).lower()}")
            elif isinstance(v, (int, float)):
                hash_parts.append(f"{k}:{v}")
            else:
                hash_parts.append(f"{k}:{v}")
        hash_string3 = '|'.join(hash_parts) + '|' + self.secure_key
        hash3 = hashlib.md5(hash_string3.encode()).hexdigest()
        
        # Method 4: Simple concatenation of values in sorted key order
        hash_string4 = ''.join(str(payload_for_hash[k]) for k in sorted_keys) + self.secure_key
        hash4 = hashlib.md5(hash_string4.encode()).hexdigest()
        
        # Method 5: Key fields only (most important ones)
        key_fields = [
            payload.get('searchWord', ''),
            str(payload.get('shopId', 0)),
            payload.get('deviceId', ''),
            payload.get('deviceToken', ''),
            str(payload.get('offset', 0)),
            str(payload.get('limit', 24)),
            str(payload.get('pageNo', 1))
        ]
        hash_string5 = ''.join(key_fields) + self.secure_key
        hash5 = hashlib.md5(hash_string5.encode()).hexdigest()
        
        # Use Method 1 (most common pattern: sorted JSON + securekey)
        # If this doesn't work, you may need to test which method produces
        # the hash 30c23fb01d501a7bf6d4e68019a9992d for the exact curl payload
        return hash1
    
    def _test_hash_generation(self) -> None:
        """
        Test hash generation with known payload to reverse engineer the algorithm
        This uses the exact payload from the curl command for "kurta"
        """
        test_payload = {
            "categoryId": "-1",
            "categoryName": "",
            "fp": "",
            "sorting": "popular:asc",
            "limit": 24,
            "offset": 0,
            "pageNo": 1,
            "pageName": "",
            "searchWord": "kurta",
            "requestMode": "similarproducts",
            "storeId": 0,
            "customerId": 0,
            "cartId": 0,
            "shopId": 33,
            "shopName": "Tasva",
            "regionID": "UL",
            "deviceId": "1603584e4a91d840b878021fedc187f5",
            "isS2SCall": False,
            "deviceType": "desktop",
            "brand": "Tasva",
            "geoLocation": -1,
            "fcmToken": -1,
            "deviceToken": "c7c647b4180847612266af3a9190ae4e72fe6ab36ab44cf13a936a7cc10b5cf9.1768676366",
        }
        expected_hash = "30c23fb01d501a7bf6d4e68019a9992d"
        
        # Test different methods
        payload_for_hash = test_payload.copy()
        methods = []
        
        # Method 1: Sorted JSON + securekey
        payload_json = json.dumps(payload_for_hash, sort_keys=True, separators=(',', ':'))
        hash1 = hashlib.md5((payload_json + self.secure_key).encode()).hexdigest()
        methods.append(("Sorted JSON + securekey", hash1))
        
        # Method 2: securekey + Sorted JSON
        hash2 = hashlib.md5((self.secure_key + payload_json).encode()).hexdigest()
        methods.append(("securekey + Sorted JSON", hash2))
        
        # Method 3: Just sorted JSON (no securekey)
        hash3 = hashlib.md5(payload_json.encode()).hexdigest()
        methods.append(("Sorted JSON only", hash3))
        
        # Method 4: Key fields concatenated
        key_fields = [
            test_payload['searchWord'],
            str(test_payload['shopId']),
            test_payload['deviceId'],
            test_payload['deviceToken'],
            str(test_payload['offset']),
            str(test_payload['limit'])
        ]
        hash4 = hashlib.md5((''.join(key_fields) + self.secure_key).encode()).hexdigest()
        methods.append(("Key fields + securekey", hash4))
        
        # Method 5: JSON with spaces (pretty print)
        payload_json_pretty = json.dumps(payload_for_hash, sort_keys=True, indent=2)
        hash5 = hashlib.md5((payload_json_pretty + self.secure_key).encode()).hexdigest()
        methods.append(("Pretty JSON + securekey", hash5))
        
        # Method 6: JSON with spaces but no indent
        payload_json_spaces = json.dumps(payload_for_hash, sort_keys=True)
        hash6 = hashlib.md5((payload_json_spaces + self.secure_key).encode()).hexdigest()
        methods.append(("JSON with spaces + securekey", hash6))
        
        # Method 7: Exclude certain fields
        payload_minimal = {k: v for k, v in payload_for_hash.items() 
                          if k not in ['categoryName', 'fp', 'pageName', 'shopName']}
        payload_json_min = json.dumps(payload_minimal, sort_keys=True, separators=(',', ':'))
        hash7 = hashlib.md5((payload_json_min + self.secure_key).encode()).hexdigest()
        methods.append(("Minimal fields JSON + securekey", hash7))
        
        # Method 8: All fields as string values concatenated
        sorted_keys = sorted(payload_for_hash.keys())
        all_values = ''.join(str(payload_for_hash[k]) for k in sorted_keys)
        hash8 = hashlib.md5((all_values + self.secure_key).encode()).hexdigest()
        methods.append(("All values concatenated + securekey", hash8))
        
        # Method 9: HMAC-SHA256 style (but with MD5)
        import hmac
        hash9 = hmac.new(self.secure_key.encode(), payload_json.encode(), hashlib.md5).hexdigest()
        methods.append(("HMAC-MD5 (securekey as key)", hash9))
        
        # Method 10: Just the search word + securekey (simplest)
        hash10 = hashlib.md5((test_payload['searchWord'] + self.secure_key).encode()).hexdigest()
        methods.append(("searchWord + securekey", hash10))
        
        print(f"\n{'='*60}")
        print("HASH GENERATION TEST")
        print(f"{'='*60}")
        print(f"Expected hash: {expected_hash}")
        print(f"\nTested methods:")
        for method_name, generated_hash in methods:
            match = "✓ MATCH!" if generated_hash == expected_hash else "✗"
            print(f"  {match} {method_name}: {generated_hash}")
        print(f"{'='*60}\n")
    
    def search(self, query: str, offset: int = 0, limit: int = 100, 
               category_id: str = "-1", sorting: str = "popular:asc") -> Dict[str, Any]:
        """
        Search for products on Tasva
        
        Args:
            query: Search query string
            offset: Offset for pagination (default: 0)
            limit: Number of results per page (default: 100, increased to reduce API calls)
            category_id: Category ID filter (default: "-1" for all)
            sorting: Sort order (default: "popular:asc")
            
        Returns:
            JSON response from the API
        """
        # Build request payload - matching exact structure from curl command
        payload = {
            "categoryId": category_id,
            "categoryName": "",
            "fp": "",
            "sorting": sorting,
            "limit": limit,
            "offset": offset,
            "pageNo": (offset // limit) + 1,
            "pageName": "",
            "searchWord": query,
            "requestMode": "similarproducts",
            "storeId": 0,
            "customerId": 0,
            "cartId": 0,
            "shopId": 33,
            "shopName": "Tasva",
            "regionID": "UL",
            "deviceId": self.device_id,  # Use fixed deviceId from curl
            "isS2SCall": False,
            "deviceType": "desktop",
            "brand": "Tasva",
            "geoLocation": -1,
            "fcmToken": -1,
            "deviceToken": self.device_token,
        }
        
        # Try to use known working hash for specific queries
        # NOTE: The hash generation algorithm hasn't been fully reverse-engineered.
        # For queries/offsets not in this dict, the scraper will generate a hash
        # that may not work. You can add more known hashes here by capturing them
        # from browser network requests.
        # Known hash for "kurta" at offset 0: 30c23fb01d501a7bf6d4e68019a9992d
        known_hashes = {
            ("kurta", 0): "30c23fb01d501a7bf6d4e68019a9992d",
            # Add more known hashes here as you discover them:
            # ("sherwani", 0): "your_hash_here",
            # ("bundi", 0): "your_hash_here",
        }
        
        if (query, offset) in known_hashes:
            payload["hash"] = known_hashes[(query, offset)]
            payload["validateHash"] = True
            print(f"Using known working hash for '{query}' at offset {offset}")
        else:
            # Generate hash (must be done after all fields are set)
            payload["hash"] = self._generate_hash(payload)
            payload["validateHash"] = True
            print(f"GENERATED HASH: {payload['hash']}")
            # Try without hash validation as fallback
            # We'll try with validateHash first, then without if it fails
        
        # Debug: Print hash for troubleshooting
        if query == "kurta" and offset == 0:
            print(f"EXPECTED HASH (for reference): 30c23fb01d501a7bf6d4e68019a9992d")
        
        print(f"\n{'='*60}")
        print(f"REQUEST URL: {self.BASE_URL}")
        print(f"QUERY: '{query}' | OFFSET: {offset} | LIMIT: {limit}")
        print(f"{'='*60}")
        
        try:
            response = self.session.post(self.BASE_URL, json=payload, timeout=30)
            
            # Log response details
            print(f"\nRESPONSE STATUS CODE: {response.status_code}")
            print(f"CONTENT TYPE: {response.headers.get('Content-Type', 'Unknown')}")
            print(f"CONTENT LENGTH: {len(response.content)} bytes")
            
            # If we get 400 with INVALID_REQUEST, try without hash validation
            if response.status_code == 400:
                try:
                    error_data = response.json()
                    if isinstance(error_data, dict) and error_data.get('msg') == ["INVALID_REQUEST"]:
                        print("⚠ Got INVALID_REQUEST error, trying without hash validation...")
                        # Try again without hash validation
                        payload_no_hash = payload.copy()
                        payload_no_hash.pop('hash', None)
                        payload_no_hash['validateHash'] = False
                        response = self.session.post(self.BASE_URL, json=payload_no_hash, timeout=30)
                        print(f"Retry STATUS CODE: {response.status_code}")
                except:
                    pass
            
            response.raise_for_status()
            
            # Parse JSON response
            try:
                data = response.json()
                print(f"\n✓ Successfully parsed as JSON")
                
                # Log response structure summary
                if isinstance(data, dict):
                    print(f"Response keys: {list(data.keys())}")
                    if 'success' in data:
                        print(f"Success: {data['success']}")
                    if 'results' in data and isinstance(data['results'], dict):
                        if 'products' in data['results']:
                            products_data = data['results']['products']
                            if isinstance(products_data, dict):
                                print(f"Products keys: {list(products_data.keys())}")
                                if 'total' in products_data:
                                    print(f"Total products available: {products_data['total']}")
                                if 'hits' in products_data:
                                    print(f"Products in this page: {len(products_data['hits'])}")
                
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
        
        # Tasva API structure: {success: true, results: {products: {hits: [...]}}}
        if isinstance(data, dict) and data.get('success'):
            results = data.get('results', {})
            if isinstance(results, dict):
                products_obj = results.get('products', {})
                if isinstance(products_obj, dict):
                    hits = products_obj.get('hits', [])
                    if isinstance(hits, list):
                        print(f"Found {len(hits)} products in hits array")
                        for idx, hit in enumerate(hits):
                            print(f"  Processing product {idx+1}...")
                            # Product data is in _source field
                            if isinstance(hit, dict) and '_source' in hit:
                                product_data = hit['_source']
                                product = self._extract_product_info(product_data)
                                if product:
                                    print(f"    ✓ Extracted product: {product.get('Name', product.get('ProductID', 'unknown'))}")
                                    products.append(product)
                                else:
                                    print(f"    ✗ No product data extracted")
                            elif isinstance(hit, dict):
                                # Sometimes product data might be directly in hit
                                product = self._extract_product_info(hit)
                                if product:
                                    products.append(product)
        
        print(f"\nTotal products extracted: {len(products)}")
        print(f"{'='*60}\n")
        
        return products
    
    def _extract_product_info(self, item: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extract product information from a product item
        
        Args:
            item: Product item from API response (_source field)
            
        Returns:
            Dictionary with extracted product information
        """
        product = {}
        
        # Extract basic product fields
        product['ProductID'] = item.get('ProductID', '')
        product['StyleCode'] = item.get('StyleCode', '')
        product['Name'] = item.get('Name', '')
        product['ShortDescription'] = item.get('ShortDescription', '')
        product['LinkRewrite'] = item.get('LinkRewrite', '')
        product['Price'] = item.get('Price', '')
        product['SellingPrice'] = item.get('SellingPrice', '')
        product['Color'] = item.get('Color', '')
        product['Quantity'] = item.get('Quantity', 0)
        product['Active'] = item.get('Active', 0)
        product['IsReturnable'] = item.get('IsReturnable', 0)
        
        # Category information
        product['DefaultCategoryID'] = item.get('DefaultCategoryID', '')
        product['DefaultCategoryName'] = item.get('DefaultCategoryName', '')
        product['DefaultCategoryLinkRewrite'] = item.get('DefaultCategoryLinkRewrite', '')
        product['CategoryGroupID'] = item.get('CategoryGroupID', '')
        
        # Gender information
        product['Gender'] = item.get('Gender', '')
        product['GenderLinkRewrite'] = item.get('GenderLinkRewrite', '')
        
        # Shop information
        product['ShopID'] = item.get('ShopID', '')
        product['ShopIDs'] = json.dumps(item.get('ShopIDs', [])) if isinstance(item.get('ShopIDs'), list) else item.get('ShopIDs', '')
        
        # Dates
        product['PublishedDate'] = item.get('PublishedDate', '')
        product['FirstInStock'] = item.get('FirstInStock', '')
        
        # Features (nested object)
        features = item.get('Features', {})
        if isinstance(features, dict):
            product['Brand'] = features.get('Brand', '')
            product['Fabric'] = features.get('Fabric', '')
            product['Fit'] = features.get('Fit', '')
            product['Color_Feature'] = features.get('Color', '')
            product['Craft'] = features.get('Craft', '')
            product['Collection'] = features.get('Collection', '')
            product['Occasion'] = features.get('Occasion', '')
            product['Length'] = features.get('Length', '')
            product['Neck'] = features.get('Neck', '')
            product['SleeveLength'] = features.get('SleeveLength', '')
            product['Type'] = features.get('Type', '')
            product['Pockets'] = features.get('Pockets', '')
            product['Reversible'] = features.get('Reversible', '')
            product['Closure'] = features.get('Closure', '')
            product['Subbrand'] = features.get('Subbrand', '')
            # Store full features as JSON for reference
            product['Features_JSON'] = json.dumps(features, ensure_ascii=False)
        else:
            product['Features_JSON'] = json.dumps(features) if features else ''
        
        # Discount information
        discount = item.get('Discount', {})
        if isinstance(discount, dict):
            product['Discount_Type'] = discount.get('Type', '')
            product['Discount_Amount'] = discount.get('Amount', '')
            product['Discount_StartDate'] = discount.get('StartDate', '')
            product['Discount_EndDate'] = discount.get('EndDate', '')
            product['Discount_JSON'] = json.dumps(discount, ensure_ascii=False)
        else:
            product['Discount_JSON'] = json.dumps(discount) if discount else ''
        
        # Media (Images)
        media = item.get('Media', {})
        if isinstance(media, dict):
            images = media.get('Images', [])
            if isinstance(images, list):
                # Extract image URLs/names
                image_names = []
                image_urls = []
                for img in images:
                    if isinstance(img, dict):
                        img_name = img.get('Name', '')
                        if img_name:
                            image_names.append(img_name)
                            # Generate full image URL
                            image_url = f"https://imagescdn.tasva.com/img/app/product/1/{img_name}.jpg?w=1000&auto=format"
                            image_urls.append(image_url)
                product['Images'] = ', '.join(image_names)
                product['Images_JSON'] = json.dumps(images, ensure_ascii=False)
                # Add image URL columns
                product['ImageURLs'] = ', '.join(image_urls) if image_urls else ''
                product['ImageURLs_JSON'] = json.dumps(image_urls, ensure_ascii=False) if image_urls else '[]'
                product['PrimaryImageURL'] = image_urls[0] if image_urls else ''
            else:
                product['ImageURLs'] = ''
                product['ImageURLs_JSON'] = '[]'
                product['PrimaryImageURL'] = ''
            
            # Swatch image
            swatch = media.get('Swatch', '')
            product['Swatch'] = swatch
            product['SwatchExtension'] = media.get('SwatchExtension', '')
            # Generate swatch URL
            if swatch:
                # Remove .jpg extension if present
                swatch_clean = swatch.replace('.jpg', '').replace('.jpeg', '')
                product['SwatchImageURL'] = f"https://imagescdn.tasva.com/img/app/product/1/{swatch_clean}.jpg?w=1000&auto=format"
            else:
                product['SwatchImageURL'] = ''
        else:
            product['Images_JSON'] = json.dumps(media) if media else ''
            product['ImageURLs'] = ''
            product['ImageURLs_JSON'] = '[]'
            product['PrimaryImageURL'] = ''
            product['SwatchImageURL'] = ''
        
        # Sizes (array of size objects)
        sizes = item.get('Sizes', [])
        if isinstance(sizes, list):
            size_names = []
            size_info = []
            for size in sizes:
                if isinstance(size, dict):
                    size_name = size.get('Name', '')
                    if size_name:
                        size_names.append(size_name)
                    size_info.append({
                        'Name': size.get('Name', ''),
                        'Quantity': size.get('Quantity', 0),
                        'Price': size.get('Price', ''),
                        'SellingPrice': size.get('SellingPrice', ''),
                        'ID': size.get('ID', '')
                    })
            product['Sizes'] = ', '.join(size_names)
            product['Sizes_JSON'] = json.dumps(sizes, ensure_ascii=False)
        else:
            product['Sizes_JSON'] = json.dumps(sizes) if sizes else ''
        
        # Color Away Products (variants)
        color_away = item.get('ColorAwayProducts', [])
        if isinstance(color_away, list):
            product['ColorAwayProducts_JSON'] = json.dumps(color_away, ensure_ascii=False)
        else:
            product['ColorAwayProducts_JSON'] = json.dumps(color_away) if color_away else ''
        
        # Similar Products
        product['SimilarProducts'] = item.get('SimilarProducts', '')
        
        # Store IDs
        product['StoreIds'] = item.get('StoreIds', '')
        
        # Additional metadata
        product['ProductViews'] = item.get('ProductViews', 0)
        product['OrdersCount'] = item.get('OrdersCount', 0)
        
        # Build product URL
        if product.get('LinkRewrite'):
            product['ProductURL'] = f"https://www.tasva.com/p/{product['LinkRewrite']}-{product.get('ProductID', '')}.html"
        else:
            product['ProductURL'] = ''
        
        return product
    
    def remove_duplicates_by_id(self, products: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Remove duplicate products based on ProductID
        
        Args:
            products: List of product dictionaries
            
        Returns:
            List of unique products (keeps first occurrence)
        """
        seen_ids = set()
        unique_products = []
        duplicates_count = 0
        
        for product in products:
            product_id = product.get('ProductID', '')
            if product_id and product_id in seen_ids:
                duplicates_count += 1
                continue
            
            if product_id:
                seen_ids.add(product_id)
            unique_products.append(product)
        
        if duplicates_count > 0:
            print(f"✓ Removed {duplicates_count} duplicate products (based on ProductID)")
        
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
            filename = f"tasva_products_{timestamp}.csv"
        
        if not products:
            print("No products to save", file=sys.stderr)
            return filename
        
        # Remove duplicates before saving
        print(f"\n📊 Removing duplicates...")
        print(f"   Original count: {len(products)}")
        products = self.remove_duplicates_by_id(products)
        print(f"   After deduplication: {len(products)}")
        
        # Get all unique keys from all products
        all_keys = set()
        for product in products:
            all_keys.update(product.keys())
        
        # Sort keys for consistent column order
        fieldnames = sorted(all_keys)
        
        # Ensure common fields are first
        preferred_order = ['ProductID', 'StyleCode', 'Name', 'ShortDescription', 'LinkRewrite',
                          'Price', 'SellingPrice', 'Color', 'Quantity', 'ProductURL',
                          'Brand', 'Fabric', 'Fit', 'Category', 'DefaultCategoryName',
                          'Gender', 'ShopID', 'Images', 'Sizes', 'Discount_Amount',
                          'PublishedDate', 'Active', 'IsReturnable']
        
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


def scrape_query(scraper: TasvaScraper, query: str, limit: int = 100) -> List[Dict[str, Any]]:
    """
    Scrape all products for a single query with pagination
    
    Args:
        scraper: TasvaScraper instance
        query: Search query string
        limit: Number of results per page
        
    Returns:
        List of all product dictionaries for this query
    """
    print(f"\n{'='*80}")
    print(f"🔍 SEARCHING FOR: '{query}'")
    print(f"{'='*80}")
    
    offset = 0
    all_products = []
    total_count = None
    
    print(f"\n{'='*60}")
    print(f"STARTING PAGINATION - Fetching ALL products for '{query}'")
    print(f"{'='*60}\n")
    
    while True:
        print(f"\n{'='*60}")
        print(f"Query: '{query}' | Offset: {offset} | Limit: {limit}")
        if total_count is not None:
            print(f"Progress: {len(all_products)}/{total_count} products collected")
        print(f"{'='*60}")
        
        response_data = scraper.search(query, offset=offset, limit=limit)
        
        # Extract products from this page
        products = scraper.extract_products(response_data)
        
        if not products:
            print(f"No products found at offset {offset}")
            break
        
        all_products.extend(products)
        print(f"\n✓ Collected {len(products)} products from offset {offset}")
        print(f"✓ Total products collected for '{query}': {len(all_products)}")
        
        # Check pagination info from API response
        # Structure: {success: true, results: {products: {total: X, hits: [...]}}}
        has_more = False
        if isinstance(response_data, dict) and response_data.get('success'):
            results = response_data.get('results', {})
            if isinstance(results, dict):
                products_obj = results.get('products', {})
                if isinstance(products_obj, dict):
                    # Get total count
                    if 'total' in products_obj:
                        total_count = products_obj['total']
                        print(f"📊 Total products available: {total_count}")
                    
                    # Check if there are more products
                    if total_count is not None:
                        if len(all_products) < total_count:
                            has_more = True
                        else:
                            has_more = False
                            print(f"✓ Reached total count: {len(all_products)}/{total_count}")
                    elif len(products) == limit:
                        # If we got a full page, assume there might be more
                        has_more = True
                        print(f"⚠ No total count found, assuming more pages (got full page of {limit})")
        
        if not has_more:
            print(f"\n{'='*60}")
            print(f"✓ Pagination complete for '{query}' - No more products to fetch")
            print(f"{'='*60}")
            break
        
        offset += limit
        
        # Safety limit to prevent infinite loops
        if offset > 100000:
            print(f"\n⚠ Safety limit reached: Stopped at offset {offset}")
            break
        
        # Small delay to avoid rate limiting
        time.sleep(0.5)
    
    print(f"\n{'='*80}")
    print(f"✓ COMPLETED: '{query}' - Found {len(all_products)} products")
    print(f"{'='*80}\n")
    
    return all_products


def main():
    """Main function to run the scraper with multiple queries"""
    scraper = TasvaScraper()
    
    # Check for test hash flag
    if len(sys.argv) > 1 and sys.argv[1] == "--test-hash":
        scraper._test_hash_generation()
        return
    
    # Define array of search queries
    # You can modify this array or pass queries via command line
    search_queries = [
        "kurta",
        "sherwani",
        "bundi",
        "pajama"
    ]
    
    # Allow queries to be passed as command line arguments
    # If arguments provided, use them instead of default array
    if len(sys.argv) > 1:
        # Command line format: python tasva_scraper.py "query1" "query2" "query3"
        search_queries = sys.argv[1:]
    
    print(f"\n{'='*80}")
    print(f"🚀 STARTING TASVA SCRAPER")
    print(f"{'='*80}")
    print(f"📋 Total queries to process: {len(search_queries)}")
    queries_str = ', '.join(f"'{q}'" for q in search_queries)
    print(f"📋 Queries: {queries_str}")
    print(f"{'='*80}\n")
    
    limit = 100  # Increased limit to reduce API calls (was 24, now 100)
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
            import traceback
            traceback.print_exc()
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

