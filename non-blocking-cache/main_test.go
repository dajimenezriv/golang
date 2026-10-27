package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type clientGetter func(ctx context.Context, address string) (string, error)

func (c clientGetter) Get(ctx context.Context, address string) (string, error) {
	return c(ctx, address)
}

func TestGet(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls, done, canceled atomic.Int64

		start := time.Now()
		c := NewCache(clientGetter(func(ctx context.Context, address string) (string, error) {
			calls.Add(1)
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				canceled.Add(1)
				return "", ctx.Err()
			}
			done.Add(1)
			return address, nil
		}))

		_, err := c.Get(t.Context(), "address")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get() = %v, want %v", err, ErrNotFound)
		}
		synctest.Wait()
		if got := calls.Load(); got != 1 {
			t.Errorf("calls = %d, want 1", got)
		}

		time.Sleep(999 * time.Millisecond)
		_, err = c.Get(t.Context(), "address")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Get() = %v, want %v", err, ErrNotFound)
		}
		if got := done.Load(); got != 0 {
			t.Errorf("done = %d, want 0", got)
		}

		time.Sleep(time.Millisecond)
		got, err := c.Get(t.Context(), "address")
		if err != nil {
			t.Errorf("Get() = %v, want nil", err)
		}
		if got != "address" {
			t.Errorf("Get() = %q, want address", got)
		}
		if got := done.Load(); got != 1 {
			t.Errorf("done = %d, want 1", got)
		}
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Errorf("elapsed = %v, want 1s", got)
		}
	})
}
