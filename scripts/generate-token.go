package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mgsearch/pkg/auth"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, continue without it
	}

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <store-uuid> <shop-domain>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s a1b2c3d4-e5f6-7890-abcd-ef1234567890 mg-store-207095.myshopify.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nGet store UUID from database:\n")
		fmt.Fprintf(os.Stderr, "  SELECT id, shop_domain FROM stores;\n")
		fmt.Fprintf(os.Stderr, "\nOr install a store first:\n")
		fmt.Fprintf(os.Stderr, "  curl -X POST http://localhost:8080/api/auth/shopify/install \\\n")
		fmt.Fprintf(os.Stderr, "    -H 'Content-Type: application/json' \\\n")
		fmt.Fprintf(os.Stderr, "    -d '{\"shop\":\"mg-store-207095.myshopify.com\",\"access_token\":\"...\"}'\n")
		os.Exit(1)
	}

	storeID := os.Args[1]
	shop := strings.TrimSpace(os.Args[2])
	
	// Remove https:// if present
	shop = strings.TrimPrefix(shop, "https://")
	shop = strings.TrimPrefix(shop, "http://")
	shop = strings.TrimSuffix(shop, "/")
	
	signingKey := os.Getenv("JWT_SIGNING_KEY")

	if signingKey == "" {
		fmt.Fprintf(os.Stderr, "Error: JWT_SIGNING_KEY environment variable not set\n")
		fmt.Fprintf(os.Stderr, "\nTry one of these:\n")
		fmt.Fprintf(os.Stderr, "  1. Export from .env: export JWT_SIGNING_KEY=$(grep JWT_SIGNING_KEY .env | cut -d '=' -f2)\n")
		fmt.Fprintf(os.Stderr, "  2. Set directly: export JWT_SIGNING_KEY=your-key-here\n")
		fmt.Fprintf(os.Stderr, "  3. Make sure .env file exists in the project root\n")
		os.Exit(1)
	}

	token, err := auth.GenerateSessionToken(storeID, shop, []byte(signingKey), 24*time.Hour)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Session Token (valid for 24 hours):")
	fmt.Println(token)
	fmt.Fprintf(os.Stderr, "\nUse it like this:\n")
	fmt.Fprintf(os.Stderr, "  curl -X GET 'http://localhost:8080/api/stores/current' \\\n")
	fmt.Fprintf(os.Stderr, "    -H 'Authorization: Bearer %s'\n", token)
}

