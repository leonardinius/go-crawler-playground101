package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	urls := make(chan string, 1000)
	results := make(chan CrawledLinks, 1000)
	uniqueLinks := make(map[string]bool)

	file, err := os.OpenFile("urls.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) // nolint
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	var crawler *Crawler
	if os.Getenv("USE_FAKE") == "" {
		crawler = NewCrawler(NewHTTPFetcher())
	} else {
		crawler = NewCrawler(NewFakeFetcher())
	}

	go func() {
		defer func() {
			err := file.Close()
			if err != nil {
				log.Fatalf("error closing file: %v", err)
			}
		}()
		for url := range urls {
			_, err := file.WriteString(url + "\n")
			if err != nil {
				log.Fatalf("error writing to file: %v", err)
			}
		}
	}()

	go func() {
		for result := range results {
			if result.err != nil {
				fmt.Printf("err: %v\n", result.err)
				continue
			}
			for _, url := range result.urls {
				if _, ok := uniqueLinks[url]; !ok {
					uniqueLinks[url] = true
					urls <- url
				}
			}
			unique := len(uniqueLinks)
			fmt.Printf("found: depth=%d url=%q urls=%d %s, total_unique=%d\n", result.depth, result.url, len(result.urls), crawler.MetricsSummary(), unique)
		}
	}()

	crawler.Crawl("https://golang.org/", 4, results)
}
