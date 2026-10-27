package main

import (
	"context"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type mockFetcher struct {
	fetch func(context.Context, string) (string, []string, error)
}

func (m *mockFetcher) Fetch(ctx context.Context, url string) (string, []string, error) {
	return m.fetch(ctx, url)
}

type crawlResult struct {
	calls, cancelled int64
}

func runCrawl(t *testing.T, ctx context.Context, depth int, fetch func(string) (string, []string, error)) ([]string, crawlResult, []error) {
	t.Helper()

	var calls, canceled atomic.Int64

	m := &mockFetcher{}
	m.fetch = func(ctx context.Context, url string) (string, []string, error) {
		select {
		case <-ctx.Done():
			canceled.Add(1)
			return "", []string{}, ctx.Err()
		case <-time.After(time.Second):
			calls.Add(1)
			return fetch(url)
		}
	}

	rs := Crawl(ctx, "root", depth, m)
	synctest.Wait()

	urls := make([]string, 0)
	errs := make([]error, 0)
	for r := range rs {
		urls = append(urls, r.URL)
		if r.Err != nil {
			errs = append(errs, r.Err)
		}
	}
	slices.Sort(urls)

	return urls, crawlResult{calls.Load(), canceled.Load()}, errs
}

func TestCrawlFetchesUntilDepth(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		urls, got, errs := runCrawl(t, t.Context(), 3, func(url string) (string, []string, error) {
			return "body", []string{url + ":url1", url + ":url2"}, nil
		})

		wantUrls := []string{"root", "root:url1", "root:url1:url1", "root:url1:url2", "root:url2", "root:url2:url1", "root:url2:url2"}
		if !slices.Equal(urls, wantUrls) {
			t.Errorf("urls:\n got  = %v\n want = %v", urls, wantUrls)
		}
		if got := len(errs); got != 0 {
			t.Errorf("len(errors) = %d, want 0", got)
		}
		want := crawlResult{7, 0}
		if got != want {
			t.Errorf("Crawl():\n got  = %+v\n want = %+v", got, want)
		}
	})
}

func TestCrawlDoesNotRefetchesSeenURL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		urls, got, errs := runCrawl(t, t.Context(), 3, func(url string) (string, []string, error) {
			return "body", []string{"url"}, nil
		})

		wantUrls := []string{"root", "url"}
		if !slices.Equal(urls, wantUrls) {
			t.Errorf("urls:\n got  = %v\n want = %v", urls, wantUrls)
		}
		if got := len(errs); got != 0 {
			t.Errorf("len(errors) = %d, want 0", got)
		}
		want := crawlResult{2, 0}
		if got != want {
			t.Errorf("Crawl():\n got  = %+v\n want = %+v", got, want)
		}
	})
}

func TestCrawlContextCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		urls, _, errs := runCrawl(t, ctx, 10, func(url string) (string, []string, error) {
			if url == "root:url1:url1:url1:url1:url1:url1:url1" {
				cancel()
			}
			return "body", []string{url + ":url1", url + ":url2"}, nil
		})

		if got := len(urls); got < 100 || got > 500 {
			t.Errorf("len(urls) = %d, want between 100 and 500", got)
		}
		if got := len(errs); got != 0 {
			t.Errorf("len(errors) = %d, want 0", got)
		}
	})
}
