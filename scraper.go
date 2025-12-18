package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// watchKeywordsMap stores keywords in a map for O(1) lookup (case insensitive)
var watchKeywordsMap = make(map[string]bool)

// initKeywords initializes the keywords map from a file and/or default list
func initKeywords() error {
	// Try to load from keywords-news.txt file first
	file, err := os.Open("keywords-news.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			keyword := strings.TrimSpace(scanner.Text())
			if keyword != "" && !strings.HasPrefix(keyword, "#") {
				watchKeywordsMap[strings.ToLower(keyword)] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading keywords-news.txt: %w", err)
		}
		log.Printf("Loaded %d keywords from keywords-news.txt", len(watchKeywordsMap))
	}
	return nil
}

// checkKeywords checks if any of the watch keywords are found in the text (case insensitive)
// Returns a slice of found keywords
func checkKeywords(text string) []string {
	found := []string{}
	lowerText := strings.ToLower(text)

	// Check each keyword in the map
	for keyword := range watchKeywordsMap {
		if strings.Contains(lowerText, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

// runScraper executes a single scraping run
func runScraper() error {
	// Create a new collector
	c := colly.NewCollector(
		// Visit only domains
		colly.AllowedDomains("www.hltv.org", "hltv.org"),
	)

	// Before making a request - allow homepage, but only /news paths after that
	c.OnRequest(func(r *colly.Request) {
		// Remove hash fragment if present
		if r.URL.Fragment != "" {
			r.URL.Fragment = ""
		}

		// Allow the homepage (root path)
		if r.URL.Path == "/" || r.URL.Path == "" {
			fmt.Println("Visiting homepage:", r.URL.String())
			return
		}
		// Only process requests to /news paths
		if !strings.HasPrefix(r.URL.Path, "/news") {
			r.Abort()
			return
		}
		fmt.Println("Visiting", r.URL.String())
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// Resolve relative URLs to absolute URLs
		absoluteURL := e.Request.AbsoluteURL(link)

		// Parse the absolute URL to check if it starts with "/news"
		parsedURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		// Check if path starts with "/news"
		if strings.HasPrefix(parsedURL.Path, "/news") {
			// Remove hash fragment to treat URLs with different hashes as the same
			parsedURL.Fragment = ""
			// Reconstruct URL without fragment
			cleanURL := parsedURL.String()

			// Check for keywords in link text
			linkText := e.Text
			foundKeywords := checkKeywords(linkText)
			if len(foundKeywords) > 0 {
				log.Printf("🔔 SPECIAL ALERT: Found keywords [%s] in link text: %q -> %s",
					strings.Join(foundKeywords, ", "), linkText, cleanURL)
			}

			fmt.Printf("News link found: %q -> %s (cleaned: %s)\n", e.Text, link, cleanURL)
			e.Request.Visit(cleanURL)
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// On response - only process pages from /news paths
	c.OnResponse(func(r *colly.Response) {
		if strings.HasPrefix(r.Request.URL.Path, "/news") {
			fmt.Printf("Scraping news page: %s (Status: %d)\n", r.Request.URL.String(), r.StatusCode)

			// Check for keywords in the page content
			bodyText := string(r.Body)
			foundKeywords := checkKeywords(bodyText)
			if len(foundKeywords) > 0 {
				log.Printf("🔔 SPECIAL ALERT: Found keywords [%s] on page: %s",
					strings.Join(foundKeywords, ", "), r.Request.URL.String())
			}

			// Add your data extraction logic here
		}
	})

	// Start scraping from the homepage
	log.Println("Starting scraper run from hltv.org...")
	return c.Visit("https://www.hltv.org/")
}

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
