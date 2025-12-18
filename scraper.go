package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

// runScraper executes a single scraping run
func runScraper() error {
	// Create a new collector
	c := colly.NewCollector(
		// Visit only domains
		colly.AllowedDomains("www.hltv.org", "hltv.org"),
	)

	// Track URLs that came from hotmatch-box links (one level deep only)
	hotmatchBoxURLs := make(map[string]bool)

	// Helper function to normalize URLs consistently (remove fragment only)
	normalizeURL := func(u *url.URL) string {
		uCopy := *u // Make a copy to avoid modifying the original
		uCopy.Fragment = ""
		return uCopy.String()
	}

	// Before making a request - allow homepage, /news paths, and match paths
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

		// Normalize URL for comparison
		normalizedURL := normalizeURL(r.URL)

		// Allow /news paths and match paths (from hotmatch-box)
		if !strings.HasPrefix(r.URL.Path, "/news") && !strings.HasPrefix(r.URL.Path, "/matches") {
			// Check if this is a hotmatch-box URL we're tracking
			if !hotmatchBoxURLs[normalizedURL] {
				r.Abort()
				return
			}
		}
		fmt.Println("Visiting", r.URL.String())
	})

	// Handle hotmatch-box anchor elements (one level deep only)
	c.OnHTML("a.hotmatch-box[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// Resolve relative URLs to absolute URLs
		absoluteURL := e.Request.AbsoluteURL(link)

		// Parse the absolute URL
		parsedURL, err := url.Parse(absoluteURL)
		if err != nil {
			return
		}

		// Normalize URL (remove fragment and query)
		cleanURL := normalizeURL(parsedURL)

		// Mark this URL as coming from hotmatch-box (one level deep)
		hotmatchBoxURLs[cleanURL] = true

		// Check for keywords in link text (use matches keywords for hotmatch-box)
		linkText := e.Text
		foundKeywords := checkKeywords(linkText, true) // true = isMatchPage
		if len(foundKeywords) > 0 {
			log.Printf("🔔 SPECIAL ALERT: Found keywords [%s] in hotmatch-box link text: %q -> %s",
				strings.Join(foundKeywords, ", "), linkText, cleanURL)
		}

		fmt.Printf("Hotmatch-box link found: %q -> %s (cleaned: %s)\n", e.Text, link, cleanURL)
		e.Request.Visit(cleanURL)
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Skip if we're on a page that came from a hotmatch-box link (one level deep only)
		// Normalize current URL for comparison
		currentURL := normalizeURL(e.Request.URL)
		if hotmatchBoxURLs[currentURL] {
			return // Don't follow links on hotmatch-box pages
		}

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

			// Check for keywords in link text (use news keywords for news links)
			linkText := e.Text
			foundKeywords := checkKeywords(linkText, false) // false = isNewsPage
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

	// On response - process pages from /news paths and hotmatch-box pages
	c.OnResponse(func(r *colly.Response) {
		// Normalize URL for comparison
		currentURL := normalizeURL(r.Request.URL)
		isHotmatchBoxPage := hotmatchBoxURLs[currentURL]

		if strings.HasPrefix(r.Request.URL.Path, "/news") || isHotmatchBoxPage {
			pageType := "news"
			isMatchPage := false
			if isHotmatchBoxPage {
				pageType = "hotmatch-box"
				isMatchPage = true
			}
			fmt.Printf("Scraping %s page: %s (Status: %d)\n", pageType, r.Request.URL.String(), r.StatusCode)

			// Check for keywords in the page content (use appropriate keyword map)
			bodyText := string(r.Body)
			foundKeywords := checkKeywords(bodyText, isMatchPage)
			if len(foundKeywords) > 0 {
				log.Printf("🔔 SPECIAL ALERT: Found keywords [%s] on %s page: %s",
					strings.Join(foundKeywords, ", "), pageType, r.Request.URL.String())
			}

			// Add your data extraction logic here
		}
	})

	// Start scraping from the homepage
	log.Println("Starting scraper run from hltv.org...")
	return c.Visit("https://www.hltv.org/")
}
