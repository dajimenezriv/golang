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

type getWithWorkersResult struct {
	val     string
	calls   int64
	elapsed time.Duration
}

func runGetWithWorkers(t *testing.T, ctx context.Context, get func(address string) (string, error), n int) (getWithWorkersResult, error) {
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
	val, err := GetWithWorkers(ctx, m, addresses, "key")
	return getWithWorkersResult{val, calls.Load(), time.Since(start)}, err
}

func TestGetWithWorkersReturnFirstSuccessfulResponse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		got, err := runGetWithWorkers(t, t.Context(), func(address string) (string, error) {
			time.Sleep(time.Second)
			if address == "address18" {
				time.Sleep(5 * time.Millisecond)
				return "success", nil
			}
			return "", errors.New("boom")
		}, 100)

		if err != nil {
			t.Errorf("GetWithWorkers() = %v, want nil", err)
		}
		want := getWithWorkersResult{"success", 25, 4005 * time.Millisecond}
		if got != want {
			t.Errorf("GetWithWorkers():\n got  = %+v\n want = %+v", got, want)
		}
		time.Sleep(time.Second)
	})
}

func TestGetWithWorkersLookupError(t *testing.T) {
	tests := []struct {
		name      string
		addresses int
		want      getWithWorkersResult
	}{
		{"no addresses", 0, getWithWorkersResult{"", 0, 0}},
		{"all requests fail", 100, getWithWorkersResult{"", 100, 20 * time.Second}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				got, err := runGetWithWorkers(t, t.Context(), func(address string) (string, error) {
					time.Sleep(time.Second)
					return "", errors.New("boom")
				}, tt.addresses)

				if !errors.Is(err, ErrLookup) {
					t.Errorf("GetWithWorkers() = %v, want %v", err, ErrLookup)
				}
				if got != tt.want {
					t.Errorf("GetWithWorkers():\n got  = %+v\n want = %+v", got, tt.want)
				}
			})
		})
	}
}

func TestGetWithWorkersContextCanceled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := runGetWithWorkers(t, ctx, func(string) (string, error) { return "success", nil }, 100)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("GetWithWorkers() = %v, want %v", err, context.Canceled)
		}
		want := getWithWorkersResult{"", 0, 0}
		if got != want {
			t.Errorf("GetWithWorkers():\n got  = %+v\n want = %+v", got, want)
		}
	})
}
