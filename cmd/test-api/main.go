package main

import (
	"fmt"
	"log"

	"github.com/mohammadalipour/keepa/internal/adapters/scraper"
)

func main() {
	client, err := scraper.NewTLSClient()
	if err != nil {
		log.Fatal(err)
	}

	url := "https://api.digikala.com/v2/product/11346346/"
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	// Print full response
	fmt.Println(resp)
}
