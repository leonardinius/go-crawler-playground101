package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

var _ Fetcher = fakeFetcher{}

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(uri string, ch chan<- string) (body string, urls []string, err error) {
	ms, err := rand.Int(rand.Reader, big.NewInt(700))
	if err != nil {
		return "", nil, err
	}
	time.Sleep(time.Duration(ms.Int64()) * time.Millisecond)
	if res, ok := f[uri]; ok {
		for _, u := range res.urls {
			ch <- u
		}
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", uri)
}

// NewFakeFetcher returns a fakeFetcher.
func NewFakeFetcher() Fetcher {
	// fetcher is a populated fakeFetcher.
	return fakeFetcher{
		"https://golang.org/": &fakeResult{
			"The Go Programming Language",
			[]string{
				"https://golang.org/pkg/",
				"https://golang.org/cmd/",
			},
		},
		"https://golang.org/pkg/": &fakeResult{
			"Packages",
			[]string{
				"https://golang.org/",
				"https://golang.org/cmd/",
				"https://golang.org/pkg/fmt/",
				"https://golang.org/pkg/os/",
			},
		},
		"https://golang.org/pkg/fmt/": &fakeResult{
			"Package fmt",
			[]string{
				"https://golang.org/",
				"https://golang.org/pkg/",
			},
		},
		"https://golang.org/pkg/os/": &fakeResult{
			"Package os",
			[]string{
				"https://golang.org/",
				"https://golang.org/pkg/",
			},
		},
	}
}
