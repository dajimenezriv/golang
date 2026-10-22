// Package pool runs jobs on a fixed number of workers and stops at the first error.
package pool

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

var ErrInvalidWorkers = errors.New("pool: workers must be a positive number")

// Func processes one job. It must return promptly when ctx is cancelled.
type Func[J, R any] func(ctx context.Context, job J) (R, error)

// Run processes jobs on the given number of workers and passes each result
// to emit. Results arrive in no particular order. emit is always called from
// a single goroutine, so it needs no locking.
//
// Run stops at the first error from do or emit, cancels the work in
// progress, and returns that error. If ctx is cancelled, Run returns
// ctx.Err(). Run returns only after every goroutine it started has exited.
func Run[J, R any](ctx context.Context, workers int, jobs []J, do Func[J, R], emit func(R) error) error {
	if workers < 1 {
		return ErrInvalidWorkers
	}

	eg, ctx := errgroup.WithContext(ctx)
	in := make(chan J)
	out := make(chan R)

	// Producer: the only sender on in, so it closes it.
	eg.Go(func() error {
		defer close(in)
		for _, j := range jobs {
			select {
			case in <- j:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	// Workers: the only senders on out, so out is closed once all of them return.
	var senders sync.WaitGroup
	for range workers {
		senders.Add(1)
		eg.Go(func() error {
			defer senders.Done()
			for j := range in {
				if err := ctx.Err(); err != nil {
					return err
				}
				r, err := do(ctx, j)
				if err != nil {
					return err
				}
				select {
				case out <- r:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	eg.Go(func() error {
		senders.Wait()
		close(out)
		return nil
	})

	// Consumer: drains out until it is closed.
	eg.Go(func() error {
		for r := range out {
			if err := emit(r); err != nil {
				return err
			}
		}
		return nil
	})

	return eg.Wait()
}
