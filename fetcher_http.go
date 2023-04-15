package main

import (
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	html "golang.org/x/net/html"
)

type httpFetcher struct{}

var _ Fetcher = httpFetcher{}

var (
	skipPlatforms = []string{
		"facebook.com",
		"google.com",
	}
	skipExtensions = []string{
		".jpg",
		".jpeg",
		".png",
		".gif",
		".pdf",
		".doc",
		".docx",
		".xls",
		".xlsx",
		".ppt",
		".pptx",
		".zip",
		".rar",
		".7z",
		".gz",
		".tar",
		".mp3",
		".mp4",
		".avi",
		".mov",
		".wmv",
		".flv",
		".mkv",
		".m4v",
		".mpg",
		".mpeg",
		".3gp",
		".3g2",
		".pkg",
		".exe",
		".msi",
		".apk",
		".ipa",
		".dmg",
		".iso",
		".bin",
		".vcd",
		".swf",
		".fla",
	}
)

func (_f httpFetcher) Fetch(uri string, ch chan<- string) (body string, urls []string, err error) {
	parsedURI, err := neturl.Parse(uri)
	if err != nil {
		return "", nil, err
	}

	for _, platform := range skipPlatforms {
		if strings.Contains(parsedURI.Host, platform) {
			return "", nil, fmt.Errorf("skip url=%q, platform %s", uri, platform)
		}
	}

	path := parsedURI.EscapedPath()
	for _, extension := range skipExtensions {
		if strings.HasSuffix(path, extension) {
			return "", nil, fmt.Errorf("skip url=%q, extension %s", uri, extension)
		}
	}

	res, err := http.Get(uri) // #nosec
	if err != nil {
		return "", nil, err
	}

	if res.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("bad status: %s", res.Status)
	}

	// check if res has content type of html
	// if not, return
	if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		return "", nil, fmt.Errorf("not html: %s", res.Header.Get("Content-Type"))
	}

	defer res.Body.Close() // #nosec

	return parseBody(parsedURI, res.Body, ch)
}

func parseBody(parsedURI *neturl.URL, resBody io.ReadCloser, ch chan<- string) (body string, urls []string, err error) {
	bodyBytes, err := io.ReadAll(resBody)
	if err != nil {
		return "", nil, err
	}
	body = string(bodyBytes)
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return "", nil, err
	}

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}

				discovered := a.Val
				if strings.HasPrefix(discovered, "./") {
					discovered = parsedURI.Scheme + "://" + parsedURI.Host + a.Val
				}
				if strings.HasPrefix(discovered, "/") {
					discovered = parsedURI.Scheme + "://" + parsedURI.Host + a.Val
				}
				parsedDiscovered, err := neturl.Parse(discovered)
				if err != nil || !strings.HasPrefix(parsedDiscovered.Scheme, "http") {
					continue
				}

				ch <- discovered
				urls = append(urls, discovered)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return body, urls, nil
}

// NewHTTPFetcher returns a new httpFetcher.
func NewHTTPFetcher() Fetcher {
	return httpFetcher{}
}
