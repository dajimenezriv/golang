package pool

import (
	"context"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func vals(n int) []int {
	var vs []int
	for i := range n {
		vs = append(vs, i)
	}
	return vs
}

type poolResult struct {
	errs            int
	done, cancelled int64
	elapsed         time.Duration
}

func runPool(t *testing.T, jobs int, timeout time.Duration, fail int) ([]int, poolResult) {
	t.Helper()

	in := make(chan int, jobs)
	for i := range jobs {
		in <- i
	}
	close(in)

	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()

	var done, cancelled atomic.Int64
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
		done.Add(1)
		return j, nil
	})

	var vals []int
	var errs int
	for r := range out {
		if r.Err != nil {
			errs++
		} else {
			vals = append(vals, r.Val)
		}
	}

	return vals, poolResult{errs, done.Load(), cancelled.Load(), time.Since(start)}
}

func TestPoolInvalidWorkers(t *testing.T) {
	defer func() {
		if r := recover(); r != ErrInvalidWorkers {
			t.Errorf("recover() = %v, want %v", r, ErrInvalidWorkers)
		}
	}()

	Pool(t.Context(), 0, nil, func(context.Context, int) (int, error) { return 0, nil })
}

func TestPoolProcessAllJobs(t *testing.T) {
	tests := []struct {
		name     string
		jobs     int
		timeout  time.Duration
		fail     int
		wantVals []int
		want     poolResult
	}{
		{"process all jobs", 100, 30 * time.Second, -1, vals(100), poolResult{0, 100, 0, 20 * time.Second}},
		{"context canceled", 100, 4500 * time.Millisecond, -1, vals(20), poolResult{0, 20, 5, 4500 * time.Millisecond}},
		{"jobs with errors", 100, 30 * time.Second, 40, vals(99), poolResult{1, 99, 0, 20 * time.Second}},
		{"no jobs", 0, time.Second, -1, vals(0), poolResult{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				vals, got := runPool(t, tt.jobs, tt.timeout, tt.fail)

				slices.Sort(vals)
				if !slices.Equal(vals, tt.wantVals) {
					t.Errorf("vals:\n got  %v\n want %v", vals, tt.wantVals)
				}
				if got != tt.want {
					t.Errorf("results:\n got =  %v, want %v", got, tt.want)
				}
			})
		})
	}
}
