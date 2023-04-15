package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	neturl "net/url"
	"sync"

	"golang.org/x/time/rate"
)

// Crawler crawls the web starting with the given url, to a maximum of depth.
type Crawler struct {
	fetcher   Fetcher
	metrics   *Metrics
	ctx       context.Context
	cache     *sync.Map
	ratelimit *sync.Map
}

// CrawledLinks is the result of a crawl.
type CrawledLinks struct {
	depth int
	url   string
	body  string
	urls  []string
	err   error
}

var (
	globalRateLimit   = rate.NewLimiter(200, 50)
	errRateLimitError = errors.New("rate limit exceeded")
)

func newHostRateLimit() *rate.Limiter {
	return rate.NewLimiter(25, 10)
}

func rateLimitBucketKey(url string) string {
	if netURL, err := neturl.Parse(url); err != nil {
		return netURL.Host
	}
	return "global"
}

// NewCrawler creates a new crawler
func NewCrawler(fetcher Fetcher) *Crawler {
	return &Crawler{
		fetcher:   fetcher,
		metrics:   NewMetrics(),
		ctx:       context.Background(),
		cache:     new(sync.Map),
		ratelimit: new(sync.Map),
	}
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func (c *Crawler) Crawl(url string, depth int, results chan<- CrawledLinks) {
	quit := make(chan bool, 1)

	go c.crawl(url, depth, quit, results)

	<-quit
	close(results)
}

func (c *Crawler) crawl(url string, depth int, quit chan bool, results chan<- CrawledLinks) {
	if depth <= 0 {
		quit <- true
		return
	}

	if _, cachePresent := c.cache.LoadOrStore(url, false); cachePresent {
		quit <- true
		return
	}

	var rl *rate.Limiter
	rlKey := rateLimitBucketKey(url)
	if _rl, rlPresent := c.ratelimit.LoadOrStore(rlKey, globalRateLimit); !rlPresent {
		rl = newHostRateLimit()
		c.ratelimit.Store(rlKey, rl)
	} else {
		rl = _rl.(*rate.Limiter)
	}

	ch := make(chan string, 5)
	wg := new(sync.WaitGroup)

	go func() {
		if err := rl.Wait(c.ctx); err != nil {
			results <- CrawledLinks{depth, url, "", nil, errRateLimitError}
			log.Fatalf("rate limit exceeded: %s, %v", url, err)
			return
		}
		if err := globalRateLimit.Wait(c.ctx); err != nil {
			results <- CrawledLinks{depth, url, "", nil, errRateLimitError}
			log.Fatalf("rate limit exceeded: %s, %v", url, err)
			return
		}
		c.metrics.IncInflightRequests()
		defer c.metrics.DecInflightRequests()
		defer c.cache.Store(url, true)
		defer close(ch)
		fmt.Printf(">> fetch %s\n", url)
		body, urls, err := c.fetcher.Fetch(url, ch)
		if err != nil {
			c.metrics.IncErrors()
		}
		c.metrics.IncReportedTotal(int64(len(urls)))
		c.metrics.IncVisitedTotal()
		results <- CrawledLinks{depth, url, body, urls, err}
	}()

	for childURL := range ch {
		wg.Add(1)
		childURLCopied := childURL
		go func() {
			defer wg.Done()
			childQuit := make(chan bool, 1)
			go c.crawl(childURLCopied, depth-1, childQuit, results)
			<-childQuit
		}()
	}
	wg.Wait()

	quit <- true
}

// MetricsSummary returns a summary of the crawler's metrics.
func (c *Crawler) MetricsSummary() string {
	return c.metrics.Summary()
}
