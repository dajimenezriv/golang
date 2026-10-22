package pool

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestPoolInvalidWorkers(t *testing.T) {
	for _, workers := range []int{-1, 0} {
		t.Run(fmt.Sprint(workers), func(t *testing.T) {
			defer func() {
				if r := recover(); r != ErrInvalidWorkers {
					t.Errorf("recover() = %v, want %v", r, ErrInvalidWorkers)
				}
			}()

			Pool(t.Context(), workers, nil, func(context.Context, int) (int, error) {
				return 0, nil
			})
		})
	}
}

func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

type poolResult struct {
	cancelled int64
	elapsed   time.Duration
}

func runPool(t *testing.T, jobs int, timeout time.Duration, fail int) ([]int, []error, poolResult) {
	t.Helper()

	in := make(chan int, jobs)
	for i := range jobs {
		in <- i
	}
	close(in)

	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()

	var cancelled atomic.Int64
	start := time.Now()
	out := Pool(ctx, 5, in, func(ctx context.Context, j int) (int, error) {
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			cancelled.Add(1)
			return 0, ctx.Err()
		}
		if j == fail {
			return 0, errBoom
		}
		return j, nil
	})

	var vals []int
	var errs []error
	for r := range out {
		if r.Err != nil {
			errs = append(errs, r.Err)
		} else {
			vals = append(vals, r.Val)
		}
	}
	slices.Sort(vals)

	return vals, errs, poolResult{cancelled.Load(), time.Since(start)}
}

func TestPool(t *testing.T) {
	tests := map[string]struct {
		jobs     int
		timeout  time.Duration
		fail     int
		wantVals []int
		wantErrs int
		want     poolResult
	}{
		"process all jobs": {
			jobs:     100,
			timeout:  30 * time.Second,
			fail:     -1,
			wantVals: seq(100),
			want:     poolResult{0, 20 * time.Second},
		},
		"context cancelled": {
			jobs:     100,
			timeout:  4500 * time.Millisecond,
			fail:     -1,
			wantVals: seq(20),
			want:     poolResult{5, 4500 * time.Millisecond},
		},
		"jobs with errors": {
			jobs:     100,
			timeout:  30 * time.Second,
			fail:     40,
			wantVals: slices.Delete(seq(100), 40, 41),
			wantErrs: 1,
			want:     poolResult{0, 20 * time.Second},
		},
		"no jobs": {
			jobs:     0,
			fail:     -1,
			wantVals: seq(0),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				vals, errs, got := runPool(t, tt.jobs, tt.timeout, tt.fail)

				if !slices.Equal(vals, tt.wantVals) {
					t.Errorf("vals:\n got  %v\n want %v", vals, tt.wantVals)
				}
				if got := len(errs); got != tt.wantErrs {
					t.Errorf("len(errs): %d, want %d", got, tt.wantErrs)
				}
				if got != tt.want {
					t.Errorf("results:\n got =  %v, want %v", got, tt.want)
				}

				for _, err := range errs {
					if !errors.Is(err, errBoom) {
						t.Errorf("error = %v, want %v", err, errBoom)
					}
				}
			})
		})
	}
}
