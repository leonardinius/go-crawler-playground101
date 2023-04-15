package main

// Fetcher - HTTP crawler, sync.
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string, ch chan<- string) (body string, urls []string, err error)
}
