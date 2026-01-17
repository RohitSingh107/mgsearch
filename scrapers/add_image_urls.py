#!/usr/bin/env python3
"""
Add complete image URLs to Tasva products CSV
Converts image IDs to full URLs using the pattern:
https://imagescdn.tasva.com/img/app/product/1/{image_id}.jpg?w=1000&auto=format
"""

import csv
import json
import sys
import os
from typing import List, Optional


def generate_image_url(image_id: str) -> str:
    """
    Generate full image URL from image ID
    
    Args:
        image_id: Image ID (e.g., "1078050-16668879")
        
    Returns:
        Complete image URL
    """
    return f"https://imagescdn.tasva.com/img/app/product/1/{image_id}.jpg?w=1000&auto=format"


def parse_images(images_str: str) -> List[str]:
    """
    Parse comma-separated image names from Images column
    
    Args:
        images_str: Comma-separated image names (e.g., "1078050-16668879, 1078050-16668881")
        
    Returns:
        List of image IDs
    """
    if not images_str or images_str.strip() == '':
        return []
    
    # Split by comma and strip whitespace
    image_ids = [img.strip() for img in images_str.split(',') if img.strip()]
    return image_ids


def parse_images_from_json(images_json_str: str) -> List[str]:
    """
    Parse image names from Images_JSON column
    
    Args:
        images_json_str: JSON string containing image array
        
    Returns:
        List of image IDs
    """
    if not images_json_str or images_json_str.strip() == '':
        return []
    
    try:
        images_data = json.loads(images_json_str)
        if isinstance(images_data, list):
            image_ids = []
            for img in images_data:
                if isinstance(img, dict) and 'Name' in img:
                    image_ids.append(img['Name'])
            return image_ids
    except (json.JSONDecodeError, TypeError):
        pass
    
    return []


def process_csv(input_file: str, output_file: Optional[str] = None) -> str:
    """
    Process CSV file and add image URL columns
    
    Args:
        input_file: Path to input CSV file
        output_file: Path to output CSV file (if None, creates new file with _with_urls suffix)
        
    Returns:
        Path to output file
    """
    if not os.path.exists(input_file):
        print(f"Error: File not found: {input_file}", file=sys.stderr)
        sys.exit(1)
    
    if output_file is None:
        # Create output filename by adding _with_urls before .csv
        base_name = os.path.splitext(input_file)[0]
        output_file = f"{base_name}_with_urls.csv"
    
    print(f"Processing: {input_file}")
    print(f"Output: {output_file}")
    
    rows_processed = 0
    rows_with_images = 0
    
    with open(input_file, 'r', encoding='utf-8') as infile, \
         open(output_file, 'w', newline='', encoding='utf-8') as outfile:
        
        reader = csv.DictReader(infile)
        fieldnames = list(reader.fieldnames)
        
        # Add new columns for image URLs
        new_columns = [
            'ImageURLs',  # Comma-separated list of all image URLs
            'ImageURLs_JSON',  # JSON array of all image URLs
            'PrimaryImageURL',  # First/main image URL
            'SwatchImageURL'  # Swatch image URL
        ]
        
        # Insert new columns after Images column
        if 'Images' in fieldnames:
            images_idx = fieldnames.index('Images')
            fieldnames = fieldnames[:images_idx+1] + new_columns + fieldnames[images_idx+1:]
        else:
            fieldnames.extend(new_columns)
        
        writer = csv.DictWriter(outfile, fieldnames=fieldnames)
        writer.writeheader()
        
        for row in reader:
            rows_processed += 1
            
            # Get image IDs from Images column
            images_str = row.get('Images', '')
            image_ids = parse_images(images_str)
            
            # If Images column is empty, try Images_JSON
            if not image_ids:
                images_json_str = row.get('Images_JSON', '')
                image_ids = parse_images_from_json(images_json_str)
            
            # Generate image URLs
            image_urls = [generate_image_url(img_id) for img_id in image_ids]
            
            # Get swatch image
            swatch_id = row.get('Swatch', '')
            swatch_url = ''
            if swatch_id:
                # Swatch might already have .jpg extension, remove it if present
                swatch_id_clean = swatch_id.replace('.jpg', '').replace('.jpeg', '')
                swatch_url = generate_image_url(swatch_id_clean)
            
            # Add new columns
            row['ImageURLs'] = ', '.join(image_urls) if image_urls else ''
            row['ImageURLs_JSON'] = json.dumps(image_urls) if image_urls else '[]'
            row['PrimaryImageURL'] = image_urls[0] if image_urls else ''
            row['SwatchImageURL'] = swatch_url
            
            if image_urls:
                rows_with_images += 1
            
            writer.writerow(row)
    
    print(f"\n✓ Processed {rows_processed} rows")
    print(f"✓ {rows_with_images} rows have images")
    print(f"✓ Saved to: {output_file}")
    
    return output_file


def main():
    """Main function"""
    if len(sys.argv) < 2:
        print("Usage: python add_image_urls.py <input_csv_file> [output_csv_file]")
        print("\nExample:")
        print("  python add_image_urls.py tasva_products_20260118_004733.csv")
        print("  python add_image_urls.py tasva_products.csv output_with_urls.csv")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None
    
    try:
        output_path = process_csv(input_file, output_file)
        print(f"\n🎉 Success! Image URLs added to: {output_path}")
    except Exception as e:
        print(f"\n✗ Error: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()

