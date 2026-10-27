package main

import (
	"context"
	"sync"
)

const maxConcurrency = 8

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

// Crawl uses fetcher to crawl pages starting with url, to a maximum of depth.
// It crawls one level at a time, so each URL is fetched once,
// at the shallowest depth it is reachable from.
// The returned channel is closed when the crawl finishes.
// Callers must drain it or cancel ctx; otherwise goroutines leak.
func Crawl(ctx context.Context, url string, depth int, fetcher Fetcher) <-chan Result {
	results := make(chan Result)

	go func() {
		defer close(results)

		var (
			mu    sync.Mutex
			seen  = map[string]struct{}{url: {}}
			sem   = make(chan struct{}, maxConcurrency)
			level = []string{url}
		)

		for range depth {
			if len(level) == 0 {
				return
			}

			var wg sync.WaitGroup
			curr := level
			level = nil
			for _, url := range curr {
				wg.Go(func() {
					select {
					case sem <- struct{}{}:
					case <-ctx.Done():
						return
					}
					body, links, err := fetcher.Fetch(ctx, url)
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

					mu.Lock()
					for _, link := range links {
						if _, ok := seen[link]; !ok {
							seen[link] = struct{}{}
							level = append(level, link)
						}
					}
					mu.Unlock()
				})
			}
			wg.Wait()
		}
	}()

	return results
}
