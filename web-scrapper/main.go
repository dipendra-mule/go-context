package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Scrapper struct {
	websites []string
}

func (s *Scrapper) ScrapeAll(ctx context.Context) map[string]string {
	var wg sync.WaitGroup
	results := make(map[string]string)

	// when we use single map and multiple workers to update the map concurrently
	// we need to use a mutex to protect the map from concurrent access
	// the lock from mutex gives access to single worker at a time to update the map
	var mu sync.Mutex

	for _, website := range s.websites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()

			// scrape with context
			data, err := s.scrapeWebsite(ctx, site)
			if err != nil {
				fmt.Printf("Error scraping %s: %v\n", site, err)
				return
			}
			mu.Lock()
			results[site] = data
			mu.Unlock()
		}(website)
	}

	wg.Wait()
	return results
}

func (s *Scrapper) scrapeWebsite(ctx context.Context, url string) (string, error) {
	// Simulate scraping with different times
	var scrapeTime time.Duration
	switch url {
	case "https://www.fast.com":
		scrapeTime = 5 * time.Second
	case "https://www.slow.com":
		scrapeTime = 10 * time.Second // this will be cancelled
	default:
		scrapeTime = 2 * time.Second
	}

	select {
	case <-time.After(scrapeTime):
		return "Data fetched from website", nil
	case <-ctx.Done():
		return "", fmt.Errorf("website scraping cancelled: %v", url)
	}
}

func main() {
	scraper := &Scrapper{
		websites: []string{
			"https://www.fast.com",
			"https://www.slow.com",
			"https://www.medium.com",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	reuslts := scraper.ScrapeAll(ctx)
	fmt.Println("Succesfull scrapes:")
	for site, data := range reuslts {
		fmt.Printf("%s: %s\n", site, data)
	}
}
