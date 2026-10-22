package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type stats struct {
	fetched, done, cancelled int64
	elapsed                  time.Duration
}

func runTest(t *testing.T, stop, drain, handle time.Duration) (stats, error) {
	t.Helper()
	var fetched, done, cancelled atomic.Int64

	ctx, cancel := context.WithTimeout(t.Context(), stop)
	defer cancel()

	start := time.Now()
	err := Run(ctx, 5, drain, func(ctx context.Context) (Message, error) {
		return Message{ID: int(fetched.Add(1))}, nil
	}, func(ctx context.Context, m Message) error {
		select {
		case <-time.After(handle):
			done.Add(1)
			return nil
		case <-ctx.Done():
			cancelled.Add(1)
			return ctx.Err()
		}
	})

	return stats{fetched.Load(), done.Load(), cancelled.Load(), time.Since(start)}, err
}

func TestRun(t *testing.T) {
	tests := []struct {
		name                string
		stop, drain, handle time.Duration
		wantErr             error
		want                stats
	}{
		{
			name:   "drains cleanly",
			stop:   2500 * time.Millisecond,
			drain:  5 * time.Second,
			handle: time.Second,
			want:   stats{fetched: 15, done: 15, elapsed: 3 * time.Second},
		},
		{
			name:    "drain timeout",
			stop:    time.Second,
			drain:   5 * time.Second,
			handle:  30 * time.Second,
			want:    stats{fetched: 5, cancelled: 5, elapsed: 6 * time.Second},
			wantErr: ErrDrainTimeout,
		},
		{
			name:   "stop before fetching",
			stop:   0,
			drain:  5 * time.Second,
			handle: time.Second,
			want:   stats{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				got, err := runTest(t, tt.stop, tt.drain, tt.handle)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Run() error = %v, want %v", err, tt.wantErr)
				}
				if got != tt.want {
					t.Errorf("Run() stats:\n got  %+v\n want %+v", got, tt.want)
				}
			})
		})
	}
}
