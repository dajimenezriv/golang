package main

import (
	"context"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(ctx context.Context, url string) (body string, urls []string, err error)
}

type Result struct {
	URL  string
	Body string
	Err  error
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(ctx context.Context, url string, depth int, fetcher Fetcher) <-chan Result {
	var (
		seen    sync.Map
		wg      sync.WaitGroup
		results = make(chan Result)
		sem     = make(chan struct{}, 8)
	)

	var visit func(url string, depth int)
	visit = func(url string, depth int) {
		if depth <= 0 {
			return
		}
		if _, loaded := seen.LoadOrStore(url, struct{}{}); loaded {
			return
		}
		wg.Go(func() {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			body, urls, err := fetcher.Fetch(ctx, url)
			<-sem

			if ctx.Err() != nil {
				return
			}
			select {
			case results <- Result{url, body, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
			for _, url := range urls {
				visit(url, depth-1)
			}
		})
	}

	visit(url, depth)

	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}
