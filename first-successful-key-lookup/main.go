package main

import (
	"context"
	"errors"
	"sync"
)

type Getter interface {
	Get(ctx context.Context, address, key string) (string, error)
}

var ErrLookup = errors.New("lookup: cannot get result, want nil")

// Call `Getter.Get()` for each address in parallel.
// Returns the first successful response.
// If all requests fail, returns an error.
func Get(ctx context.Context, getter Getter, addresses []string, key string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cAddr := make(chan string, len(addresses))
	for _, address := range addresses {
		cAddr <- address
	}
	close(cAddr)

	res := make(chan string)
	done := make(chan struct{}, 1)

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			for address := range cAddr {
				if ctx.Err() != nil {
					return
				}
				r, err := getter.Get(ctx, address, key)
				if err != nil {
					continue
				}
				select {
				case <-ctx.Done():
				case res <- r:
				}
				return
			}
		})
	}
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case r := <-res:
		return r, nil
	case <-done:
		return "", ErrLookup
	}
}
