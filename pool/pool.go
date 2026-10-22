package pool

import (
	"context"
	"sync"
)

type Result[R any] struct {
	Val R
	Err error
}

// The producer must close jobs.
func Pool[J, R any](ctx context.Context, workers int, jobs <-chan J, do func(context.Context, J) (R, error)) <-chan Result[R] {
	if workers < 1 {
		panic(ErrInvalidWorkers)
	}

	out := make(chan Result[R])
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for j := range jobs {
				r, err := do(ctx, j)
				if ctx.Err() != nil {
					return
				}

				select {
				case out <- Result[R]{r, err}:
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
