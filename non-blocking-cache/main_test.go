package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type mockClient struct {
	get                   func(ctx context.Context, address string) (string, error)
	calls, done, canceled atomic.Int64
	start                 time.Time
}

func newMockClient() *mockClient {
	return &mockClient{
		start: time.Now(),
	}
}

func (m *mockClient) Get(ctx context.Context, address string) (string, error) {
	m.calls.Add(1)
	select {
	case <-time.After(time.Second):
	case <-ctx.Done():
		m.canceled.Add(1)
		return "", ctx.Err()
	}
	m.done.Add(1)
	return address, nil
}

type getResults struct {
	v                     string
	calls, done, canceled int64
	elapsed               time.Duration
}

func (m *mockClient) collect() getResults {
	return getResults{
		calls:    m.calls.Load(),
		done:     m.done.Load(),
		canceled: m.canceled.Load(),
		elapsed:  time.Since(m.start),
	}
}

func TestGet(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		c := NewCache(m)

		_, err := c.Get(t.Context(), "address")
		if !errors.Is(err, ErrNotReady) {
			t.Errorf("Get() = %v, want %v", err, ErrNotReady)
		}
		synctest.Wait()
		want := getResults{calls: 1, done: 0, canceled: 0, elapsed: 0}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}

		// Problem, if Fatalf then it says goroutines blocked because the context from the
		// Get is outside test context
		time.Sleep(999 * time.Millisecond)
		_, err = c.Get(t.Context(), "address")
		if !errors.Is(err, ErrNotReady) {
			t.Errorf("Get() = %v, want %v", err, ErrNotReady)
		}
		want = getResults{calls: 1, done: 0, canceled: 0, elapsed: 999 * time.Millisecond}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}

		time.Sleep(time.Millisecond)
		v, err := c.Get(t.Context(), "address")
		if err != nil {
			t.Errorf("Get() = %v, want nil", err)
		}
		if v != "address" {
			t.Errorf("Get() = %q, want address", v)
		}
		want = getResults{calls: 1, done: 1, canceled: 0, elapsed: time.Second}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}
	})
}
