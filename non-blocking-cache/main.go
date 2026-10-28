// Implement a non-blocking cache that stores previously fetched responses from URLs.
// - Return cached data if the requested URL has been fetched before.
// - Fetch and store new data asynchronously if the URL is not in the cache.
// - Ensure multiple concurrent requests for the same URL do not trigger multiple fetches.
package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotReady = errors.New("cache: value not ready")

type Client interface {
	Get(ctx context.Context, address string) (string, error)
}

type Result struct {
	Val string
	Err error
}

type Cache struct {
	client Client

	mu      sync.Mutex
	flight  map[string]struct{}
	entries map[string]Result
}

func NewCache(client Client) *Cache {
	return &Cache{
		client:  client,
		flight:  make(map[string]struct{}),
		entries: make(map[string]Result),
	}
}

func (c *Cache) Get(ctx context.Context, address string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[address]; ok {
		return e.Val, e.Err
	}
	if _, ok := c.flight[address]; !ok {
		c.flight[address] = struct{}{}
		go func() {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			val, err := c.client.Get(ctx, address)
			c.mu.Lock()
			c.entries[address] = Result{val, err}
			delete(c.flight, address)
			c.mu.Unlock()
		}()
	}

	return "", ErrNotReady
}
