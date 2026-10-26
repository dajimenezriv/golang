package main

import (
	"context"

	"golang.org/x/tour/tree"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(ctx context.Context, t *tree.Tree, ch chan<- int) {
	walk(ctx, t, ch)
	close(ch)
}

func walk(ctx context.Context, t *tree.Tree, ch chan<- int) {
	if t == nil || ctx.Err() != nil {
		return
	}
	walk(ctx, t.Left, ch)
	select {
	case ch <- t.Value:
	case <-ctx.Done():
		return
	}
	walk(ctx, t.Right, ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(ctx context.Context, t1, t2 *tree.Tree) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch1, ch2 := make(chan int), make(chan int)

	go Walk(ctx, t1, ch1)
	go Walk(ctx, t2, ch2)

	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2

		if err := ctx.Err(); err != nil {
			return false, err
		}
		if ok1 != ok2 || v1 != v2 {
			return false, nil
		}
		if !ok1 {
			return true, nil
		}
	}
}
