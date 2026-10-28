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

var errBoom = errors.New("boom")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type mockClient struct {
	get                   func(address string) (string, error)
	calls, done, canceled atomic.Int64
	start                 time.Time
}

func newMockClient() *mockClient {
	return &mockClient{
		get:   func(address string) (string, error) { return address, nil },
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
	return m.get(address)
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

func assertNotReady(t *testing.T, c *Cache, ctx context.Context) {
	t.Helper()
	_, err := c.Get(ctx, "address")
	if !errors.Is(err, ErrNotReady) {
		t.Errorf("Get() = %v, want %v", err, ErrNotReady)
	}
}

func assertFetched(t *testing.T, c *Cache, m *mockClient, wantV string, wantErr error) {
	t.Helper()
	time.Sleep(time.Second)
	synctest.Wait()
	v, err := c.Get(t.Context(), "address")
	if !errors.Is(err, wantErr) {
		t.Errorf("Get() = %v, want %v", err, wantErr)
	}
	got := m.collect()
	got.v = v
	want := getResults{v: wantV, calls: 1, done: 1, canceled: 0, elapsed: time.Second}
	if got != want {
		t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
	}
}

func TestGetRunsJustOneFetch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		c := NewCache(m)
		t.Cleanup(c.Close)

		for range 100 {
			go assertNotReady(t, c, t.Context())
		}

		assertFetched(t, c, m, "address", nil)
	})
}

func TestGetFetchIgnoresContextCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		c := NewCache(m)
		t.Cleanup(c.Close)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		assertNotReady(t, c, ctx)
		assertFetched(t, c, m, "address", nil)
	})
}

func TestGetStoresError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		m.get = func(address string) (string, error) { return "", errBoom }
		c := NewCache(m)
		t.Cleanup(c.Close)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		assertNotReady(t, c, ctx)
		assertFetched(t, c, m, "", errBoom)
	})
}
