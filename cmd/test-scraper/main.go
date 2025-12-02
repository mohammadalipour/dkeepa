package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
)

// This test runner supports two modes:
//   - normal: calls DigikalaScraper.ScrapeProduct and prints parsed product/priceLog
//   - raw: creates a TLSClient and performs a direct GET to the Digikala API URL so
//     you can inspect redirects, cookies and raw body (useful for debugging redirect loops)
func main() {
	var (
		raw       = flag.Bool("raw", false, "perform a raw GET via TLSClient and print the response body")
		urlFlag   = flag.String("url", "", "optional full URL to GET when using -raw (overrides constructed API URL)")
		variantID = flag.String("variant", "", "optional variant id")
		outFile   = flag.String("out", "", "optional file to write raw body to (when -raw). If empty prints to stdout")
		timeout   = flag.Int("timeout", 30, "request timeout in seconds for TLS client")
	)
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: go run cmd/test-scraper/main.go [flags] <dkp_id>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dkpID := flag.Arg(0)

	apiURL := fmt.Sprintf("https://api.digikala.com/v2/product/%s", dkpID)
	if *variantID != "" {
		apiURL += fmt.Sprintf("?variant_id=%s", *variantID)
	}

	if *raw {
		// Use TLS client directly so we can observe redirects/cookies/raw body
		client, err := scraper.NewTLSClient()
		if err != nil {
			log.Fatalf("failed to create TLS client: %v", err)
		}

		// set a short timeout via context (the TLSClient has a built-in timeout option too)
		// but the TLSClient.Get does not accept context, so rely on its internal timeout.
		target := apiURL
		if *urlFlag != "" {
			target = *urlFlag
		}

		fmt.Printf("Performing raw GET to: %s (timeout: %ds)\n", target, *timeout)
		start := time.Now()
		body, err := client.Get(target)
		elapsed := time.Since(start)
		if err != nil {
			log.Fatalf("raw GET failed (elapsed %s): %v", elapsed, err)
		}

		fmt.Printf("raw GET succeeded (elapsed %s), body length: %d bytes\n", elapsed, len(body))

		if *outFile != "" {
			if err := ioutil.WriteFile(*outFile, []byte(body), 0644); err != nil {
				log.Fatalf("failed to write body to file: %v", err)
			}
			fmt.Printf("Wrote body to: %s\n", *outFile)
			return
		}

		// Print a truncated body to stdout to avoid flooding the terminal.
		if len(body) > 20000 {
			fmt.Printf("--- BODY (first 20KB) ---\n%s\n--- TRUNCATED ---\n", body[:20000])
		} else {
			fmt.Printf("--- BODY ---\n%s\n", body)
		}

		return
	}

	fmt.Printf("Testing scraper for DKP: %s, Variant: %s\n", dkpID, *variantID)

	s, err := scraper.NewDigikalaScraper()
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	// Use a cancellable context with a reasonable timeout for the scrape
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	product, priceLog, err := s.ScrapeProduct(ctx, dkpID, *variantID)
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	fmt.Printf("Success!\nProduct: %+v\nPriceLog: %+v\n", product, priceLog)
}
