package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type mockGetter struct {
	get func(address string) (string, error)
}

func (m *mockGetter) Get(ctx context.Context, address, key string) (string, error) {
	return m.get(address)
}

func runGet(t *testing.T, ctx context.Context, get func(address string) (string, error), n int) (string, error, int64, time.Duration) {
	t.Helper()

	var calls atomic.Int64

	m := &mockGetter{}
	m.get = func(address string) (string, error) {
		calls.Add(1)
		return get(address)
	}
	addresses := make([]string, n)
	for i := range n {
		addresses[i] = fmt.Sprintf("address%d", i)
	}

	start := time.Now()
	r, err := Get(ctx, m, addresses, "key")
	return r, err, calls.Load(), time.Since(start)
}

func TestGetReturnFirstSuccessfulResponse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, err, calls, elapsed := runGet(t, t.Context(), func(address string) (string, error) {
			time.Sleep(time.Second)
			if address == "address18" {
				time.Sleep(time.Millisecond)
				return "success", nil
			}
			return "", errors.New("boom")
		}, 100)

		if err != nil {
			t.Errorf("Get() = %v, want nil", err)
		}
		if r != "success" {
			t.Errorf("Get() = %q, want success", r)
		}
		if calls != 24 {
			t.Errorf("calls = %d, want 24", calls)
		}
		if elapsed != 4001*time.Millisecond {
			t.Errorf("elapsed = %v, want 4001ms", elapsed)
		}
		time.Sleep(time.Second)
	})
}

func TestGetLookupError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	tests := []struct {
		name        string
		ctx         context.Context
		addresses   int
		wantCalls   int64
		wantElapsed time.Duration
	}{
		{"no addresses", t.Context(), 0, 0, 0},
		{"context canceled", ctx, 100, 0, 0},
		{"all requests fail", t.Context(), 100, 100, 20 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				r, err, calls, elapsed := runGet(t, tt.ctx, func(address string) (string, error) {
					time.Sleep(time.Second)
					return "", errors.New("boom")
				}, tt.addresses)

				if !errors.Is(err, ErrLookup) {
					t.Errorf("Get() = %v, want %v", err, ErrLookup)
				}
				if r != "" {
					t.Errorf("Get() = %q, want empty", r)
				}
				if calls != tt.wantCalls {
					t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
				}
				if elapsed != tt.wantElapsed {
					t.Errorf("elapsed = %v, want %d", elapsed, tt.wantElapsed)
				}
			})
		})
	}
}
