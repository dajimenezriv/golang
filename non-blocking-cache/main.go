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

type result struct {
	val string
	err error
}

type Cache struct {
	client Client

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu      sync.Mutex
	flight  map[string]struct{}
	entries map[string]result
}

func NewCache(client Client) *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	return &Cache{
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
		flight:  make(map[string]struct{}),
		entries: make(map[string]result),
	}
}

func (c *Cache) Get(ctx context.Context, address string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[address]; ok {
		return e.val, e.err
	}
	if _, ok := c.flight[address]; !ok {
		c.flight[address] = struct{}{}
		c.wg.Go(func() {
			ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			defer cancel()
			val, err := c.client.Get(ctx, address)
			c.mu.Lock()
			c.entries[address] = result{val, err}
			delete(c.flight, address)
			c.mu.Unlock()
		})
	}

	return "", ErrNotReady
}

func (c *Cache) Close() {
	c.cancel()
	c.wg.Wait()
}
