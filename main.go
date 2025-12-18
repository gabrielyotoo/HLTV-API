package main

import (
	"log"
	"time"
)

func main() {
	// Initialize keywords from file or defaults
	if err := initKeywords(); err != nil {
		log.Fatalf("Failed to initialize keywords: %v", err)
	}

	// Run scraper immediately on startup
	log.Println("Running initial scrape...")
	if err := runScraper(); err != nil {
		log.Printf("Error during initial scrape: %v", err)
	}

	// Create a ticker that fires every 30 minutes
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	log.Println("Scraper will run every 30 minutes. Waiting for next run...")

	// Run scraper every 30 minutes
	for range ticker.C {
		log.Println("Starting scheduled scrape...")
		if err := runScraper(); err != nil {
			log.Printf("Error during scheduled scrape: %v", err)
		}
		log.Println("Scrape completed. Next run in 30 minutes...")
	}
}
