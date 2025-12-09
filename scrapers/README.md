# Snitch.com Scraper

Python scraper for extracting product data from snitch.com search API.

## Setup

1. Install Python dependencies:
```bash
pip install -r requirements.txt
```

## Usage

Run the scraper with a search query:

```bash
python snitch_scraper.py "t shirt in red"
```

If no query is provided, it defaults to "t shirt in red".

## Output

The scraper saves all product data to a CSV file named `snitch_products_YYYYMMDD_HHMMSS.csv` in the same directory.

## Features

- Uses exact headers from the provided curl command
- Extracts product information (title, price, images, URLs, etc.)
- Handles nested JSON structures
- Saves data to CSV format
- Error handling and debugging output

