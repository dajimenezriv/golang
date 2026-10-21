package pool

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

func Run2[J, R any](ctx context.Context, workers int, jobs []J, do Func[J, R], emit func(R) error) error {
	if workers < 1 {
		return ErrInvalidWorkers
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(workers)
	var mu sync.Mutex

	for _, j := range jobs {
		eg.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			r, err := do(ctx, j)
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			return emit(r)
		})
	}

	return eg.Wait()
}
