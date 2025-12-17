package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
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
			// Add your data extraction logic here
		}
	})

	// Start scraping from the homepage
	fmt.Println("Starting scraper from hltv.org...")
	c.Visit("https://www.hltv.org/")
}
