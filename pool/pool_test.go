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

func work(t *testing.T, ctx context.Context, n int, doFunc func(int) (int, error), emitFunc func(int) error) (error, int64, int, time.Duration) {
	t.Helper()

	var calls atomic.Int64
	var emitted int

	start := time.Now()
	err := Run(ctx, 5, jobs(n), func(ctx context.Context, j int) (int, error) {
		calls.Add(1)
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
		if doFunc != nil {
			return doFunc(j)
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

		err, _, _, elapsed := work(t, t.Context(), 100, nil, func(r int) error {
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
		err, calls, emitted, elapsed := work(t, t.Context(), 100, func(j int) (int, error) {
			if j == 42 {
				time.Sleep(5 * time.Millisecond)
				return 0, errBoom
			}
			return j, nil
		}, func(int) error { return nil })

		if !errors.Is(err, errBoom) {
			t.Errorf("Run() = %v, want %v", err, errBoom)
		}
		if calls != 49 {
			t.Errorf("calls = %d, want 49", calls)
		}
		if emitted != 44 {
			t.Errorf("emitted = %d, want 44", emitted)
		}
		if elapsed != 905*time.Millisecond {
			t.Errorf("elapsed = %v, want 905ms", elapsed)
		}
	})
}

func TestRunStopsWhenEmitFails(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		err, calls, emitted, elapsed := work(t, t.Context(), 100, func(j int) (int, error) {
			if j == 12 {
				time.Sleep(5 * time.Millisecond)
			}
			return j, nil
		}, func(j int) error {
			if j == 12 {
				// We wait a millisecond so the worker that sent job 12 can unblock from the
				// out channel and grab the job 19, otherwise is a race condition with the
				// cancelled context.
				time.Sleep(time.Millisecond)
				return errBoom
			}
			return nil
		})

		if !errors.Is(err, errBoom) {
			t.Errorf("Run() = %v, want %v", err, errBoom)
		}
		if calls != 20 {
			t.Errorf("calls = %d, want 20", calls)
		}
		if emitted != 15 {
			t.Errorf("emitted = %d, want 15", emitted)
		}
		if elapsed != 306*time.Millisecond {
			t.Errorf("elapsed = %v, want 306ms", elapsed)
		}
	})
}

func TestRunStopsWhenCallerCancels(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()

		err, calls, emitted, elapsed := work(t, ctx, 100, nil, func(int) error {
			return nil
		})

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Run() = %v, want %v", err, context.DeadlineExceeded)
		}
		if calls != 15 {
			t.Errorf("calls = %d, want 15", calls)
		}
		if emitted != 10 {
			t.Errorf("emitted = %d, want 10", emitted)
		}
		if elapsed != 250*time.Millisecond {
			t.Errorf("elapsed = %v, want 250ms", elapsed)
		}
	})
}

func TestRunEmptyJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		err, calls, emitted, elapsed := work(t, t.Context(), 0, nil, func(int) error {
			return nil
		})

		if err != nil {
			t.Errorf("Run() = %v, want nil", err)
		}
		if calls != 0 {
			t.Errorf("calls = %d, want 0", calls)
		}
		if emitted != 0 {
			t.Errorf("emitted = %d, want 0", emitted)
		}
		if elapsed != 0 {
			t.Errorf("elapsed = %v, want 0", elapsed)
		}
	})
}

func TestRunRejectsZeroWorkers(t *testing.T) {
	err := Run(t.Context(), 0, jobs(1), func(context.Context, int) (int, error) { return 0, nil }, func(int) error { return nil })

	if !errors.Is(err, ErrInvalidWorkers) {
		t.Errorf("Run() = %v, want %v", err, ErrInvalidWorkers)
	}
}
