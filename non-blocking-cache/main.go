// Implement a non-blocking cache that stores previously fetched responses from URLs.
// - Return cached data if the requested URL has been fetched before.
// - Fetch and store new data asynchronously if the URL is not in the cache.
// - Ensure multiple concurrent requests for the same URL do not trigger multiple fetches.
package main

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("cache: not found")

type Client interface {
	Get(ctx context.Context, address string) (string, error)
}

type Cache struct {
	client Client

	mu     sync.Mutex
	flight map[string]struct{}
	c      map[string]string
}

func NewCache(client Client) *Cache {
	return &Cache{
		client: client,
		flight: make(map[string]struct{}),
		c:      make(map[string]string),
	}
}

func (c *Cache) Get(ctx context.Context, address string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.c[address]; ok {
		return v, nil
	}
	if _, ok := c.flight[address]; !ok {
		c.flight[address] = struct{}{}
		go func() {
			defer delete(c.flight, address)
			if v, err := c.client.Get(ctx, address); err == nil {
				c.mu.Lock()
				c.c[address] = v
				c.mu.Unlock()
			}
		}()
	}

	return "", ErrNotFound
}
