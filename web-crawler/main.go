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
// The returned channel is closed when the crawl finishes.
// Callers must drain it or cancel ctx; otherwise goroutines leak.
func Crawl(ctx context.Context, url string, depth int, fetcher Fetcher) <-chan Result {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		seen    = make(map[string]int) // url -> max remaining depth it was visited with
		results = make(chan Result)
		sem     = make(chan struct{}, 8)
	)

	var visit func(url string, depth int)
	visit = func(url string, depth int) {
		if depth <= 0 {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		d, ok := seen[url]
		if ok && d >= depth {
			return
		}
		// A URL reached again with more remaining depth is fetched and
		// reported again, so a URL may appear in more than one Result.
		seen[url] = depth
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
			case results <- Result{URL: url, Body: body, Err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
			for _, u := range urls {
				visit(u, depth-1)
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
