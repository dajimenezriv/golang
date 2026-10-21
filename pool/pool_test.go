package pool

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

var errBoom = errors.New("boom")

func work(t *testing.T, n int, fail int, emitFunc func(int) error) (error, int64, int, time.Duration) {
	t.Helper()

	var calls atomic.Int64
	var emitted int

	start := time.Now()
	err := Run(t.Context(), 5, jobs(n), func(ctx context.Context, j int) (int, error) {
		calls.Add(1)
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
		if j == fail {
			time.Sleep(5 * time.Millisecond)
			return 0, errBoom
		}
		return j, nil
	}, func(r int) error {
		emitted++
		return emitFunc(r)
	})

	return err, calls.Load(), emitted, time.Since(start)
}

func jobs(n int) []int {
	js := make([]int, n)
	for i := range js {
		js[i] = i
	}
	return js
}

func TestRunSuccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var got []int

		err, _, _, elapsed := work(t, 100, -1, func(r int) error {
			got = append(got, r)
			return nil
		})

		if err != nil {
			t.Errorf("Run() = %v, want nil", err)
		}
		slices.Sort(got)
		if !slices.Equal(got, jobs(100)) {
			t.Errorf("Run().results = %v, want 0..99", got)
		}
		if elapsed != 2*time.Second {
			t.Errorf("elapsed = %v, want 2s", elapsed)
		}
	})
}

func TestRunStopsOnFirstError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		err, calls, emitted, elapsed := work(t, 100, 42, func(int) error { return nil })

		if !errors.Is(err, errBoom) {
			t.Errorf("Run() = %v, want %v", err, errBoom)
		}
		if calls != 49 {
			t.Errorf("calls = %d, want 49", calls)
		}
		if emitted != 44 {
			t.Errorf("emitted = %d, want 44", emitted)
		}
		// Job 42 runs in the ninth batch of five, so it fails at 900ms.
		if elapsed != 905*time.Millisecond {
			t.Errorf("elapsed = %v, want 905ms", elapsed)
		}
	})
}

func TestRunStopsWhenEmitFails(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		emitted := 0

		start := time.Now()
		err := Run(t.Context(), 5, jobs(100), func(ctx context.Context, j int) (int, error) {
			calls.Add(1)
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return 0, ctx.Err()
			}
			if j == 12 {
				// We want to delay job, otherwise it can arrive earlier than the other
				// goroutines, and waiting on the emit function doesn't fix it since it blocks
				// the other results.
				time.Sleep(5 * time.Millisecond)
			}
			return j, nil
		}, func(j int) error {
			emitted++
			if j == 12 {
				// We wait a millisecond so the worker that sent job 12 can unblock from the
				// out channel and grab the job 15, otherwise is a race condition with the
				// cancelled context.
				time.Sleep(time.Millisecond)
				return errBoom
			}
			return nil
		})

		if !errors.Is(err, errBoom) {
			t.Errorf("Run() = %v, want %v", err, errBoom)
		}
		if got := calls.Load(); got != 20 {
			t.Errorf("Run().calls = %d, want 20", got)
		}
		if emitted != 15 {
			t.Errorf("Run().emitted = %d, want 15", emitted)
		}
		// Job 12 runs in the third batch of five, so it fails at 300ms.
		if got, want := time.Since(start), 306*time.Millisecond; got != want {
			t.Errorf("Run().elapsed = %v, want %v", got, want)
		}
	})
}

func TestRunStopsWhenCallerCancels(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		emitted := 0
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := Run(ctx, 5, jobs(100), func(ctx context.Context, j int) (int, error) {
			calls.Add(1)
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				return 0, ctx.Err()
			}
			return j, nil
		}, func(j int) error {
			emitted++
			return nil
		})

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Run() = %v, want %v", err, context.DeadlineExceeded)
		}
		if got, want := time.Since(start), 250*time.Millisecond; got != want {
			t.Errorf("Run().elapsed = %v, want %v", got, want)
		}
	})
}

func TestRunEmptyJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int64
		emitted := 0

		err := Run(t.Context(), 5, []int{}, func(ctx context.Context, j int) (int, error) {
			calls.Add(1)
			return 0, nil
		}, func(int) error {
			emitted++
			return nil
		})

		if err != nil {
			t.Errorf("Run() = %v, want nil", err)
		}
	})
}
