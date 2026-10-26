package main

import (
	"context"
	"errors"
	"sync"
)

type Getter interface {
	Get(ctx context.Context, address, key string) (string, error)
}

var ErrLookup = errors.New("lookup: all addresses failed")

func Get(ctx context.Context, getter Getter, addresses []string, key string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val string
		err error
	}
	results := make(chan result, len(addresses))

	var wg sync.WaitGroup
	for _, address := range addresses {
		wg.Go(func() {
			val, err := getter.Get(ctx, address, key)
			results <- result{val, err}
		})
	}

	errs := []error{ErrLookup}
	for range addresses {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case r := <-results:
			if r.err == nil {
				return r.val, nil
			}
			errs = append(errs, r.err)
		}
	}
	return "", errors.Join(errs...)
}

func GetWithWorkers(ctx context.Context, getter Getter, addresses []string, key string) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string, len(addresses))
	for _, address := range addresses {
		jobs <- address
	}
	close(jobs)

	type result struct {
		val string
		err error
	}
	results := make(chan result, len(addresses))

	for range min(5, len(addresses)) {
		go func() {
			for address := range jobs {
				if ctx.Err() != nil {
					return
				}
				val, err := getter.Get(ctx, address, key)
				results <- result{val, err}
			}
		}()
	}

	errs := []error{ErrLookup}
	for range addresses {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case r := <-results:
			if r.err == nil {
				return r.val, nil
			}
			errs = append(errs, r.err)
		}
	}
	return "", errors.Join(errs...)
}
