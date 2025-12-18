package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// watchKeywordsNewsMap stores news keywords in a map for O(1) lookup (case insensitive)
var watchKeywordsNewsMap = make(map[string]bool)

// watchKeywordsMatchesMap stores match keywords in a map for O(1) lookup (case insensitive)
var watchKeywordsMatchesMap = make(map[string]bool)

// initKeywords initializes the keywords maps from files
func initKeywords() error {
	// Load news keywords
	file, err := os.Open("keywords-news.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			keyword := strings.TrimSpace(scanner.Text())
			if keyword != "" && !strings.HasPrefix(keyword, "#") {
				watchKeywordsNewsMap[strings.ToLower(keyword)] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading keywords-news.txt: %w", err)
		}
		log.Printf("Loaded %d keywords from keywords-news.txt", len(watchKeywordsNewsMap))
	}

	// Load matches keywords
	file, err = os.Open("keywords-matches.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			keyword := strings.TrimSpace(scanner.Text())
			if keyword != "" && !strings.HasPrefix(keyword, "#") {
				watchKeywordsMatchesMap[strings.ToLower(keyword)] = true
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading keywords-matches.txt: %w", err)
		}
		log.Printf("Loaded %d keywords from keywords-matches.txt", len(watchKeywordsMatchesMap))
	}

	return nil
}

// checkKeywords checks if any of the watch keywords are found in the text (case insensitive)
// Returns a slice of found keywords. Uses the appropriate keyword map based on isMatchPage.
func checkKeywords(text string, isMatchPage bool) []string {
	found := []string{}
	lowerText := strings.ToLower(text)

	// Select the appropriate keyword map
	keywordMap := watchKeywordsNewsMap
	if isMatchPage {
		keywordMap = watchKeywordsMatchesMap
	}

	// Check each keyword in the map
	for keyword := range keywordMap {
		if strings.Contains(lowerText, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}
