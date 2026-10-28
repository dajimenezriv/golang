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

func TestGetRunsJustOneFetch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		c := NewCache(m)
		t.Cleanup(c.Close)

		for range 100 {
			go func() {
				_, err := c.Get(t.Context(), "address")
				if !errors.Is(err, ErrNotReady) {
					t.Errorf("Get() = %v, want %v", err, ErrNotReady)
				}
			}()
		}

		time.Sleep(time.Second)
		synctest.Wait()
		v, err := c.Get(t.Context(), "address")
		if err != nil {
			t.Errorf("Get() = %v, want nil", err)
		}
		if v != "address" {
			t.Errorf("Get() = %q, want address", v)
		}
		want := getResults{calls: 1, done: 1, canceled: 0, elapsed: time.Second}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}
	})
}

func TestGetFetchIgnoresContextCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := newMockClient()
		c := NewCache(m)
		t.Cleanup(c.Close)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := c.Get(ctx, "address")
		if !errors.Is(err, ErrNotReady) {
			t.Errorf("Get() = %v, want %v", err, ErrNotReady)
		}

		time.Sleep(time.Second)
		synctest.Wait()
		v, err := c.Get(t.Context(), "address")
		if err != nil {
			t.Errorf("Get() = %v, want nil", err)
		}
		if v != "address" {
			t.Errorf("Get() = %q, want address", v)
		}
		want := getResults{calls: 1, done: 1, canceled: 0, elapsed: time.Second}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}
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

		_, err := c.Get(ctx, "address")
		if !errors.Is(err, ErrNotReady) {
			t.Errorf("Get() = %v, want %v", err, ErrNotReady)
		}

		time.Sleep(time.Second)
		synctest.Wait()
		_, err = c.Get(t.Context(), "address")
		if !errors.Is(err, errBoom) {
			t.Errorf("Get() = %v, want %v", err, errBoom)
		}
		want := getResults{calls: 1, done: 1, canceled: 0, elapsed: time.Second}
		if got := m.collect(); got != want {
			t.Fatalf("Get():\n got  = %+v\n want = %+v", got, want)
		}
	})
}
