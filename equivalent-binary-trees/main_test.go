package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/tour/tree"
)

func TestWalk(t *testing.T) {
	n := 10
	ch := make(chan int)
	tr := tree.New(n)
	go Walk(t.Context(), tr, ch)
	for v := range ch {
		fmt.Println(v)
	}
}

func TestSame(t *testing.T) {
	x, _ := Same(t.Context(), tree.New(10), tree.New(10))
	fmt.Println(x)
}

func TestSameContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := Same(ctx, tree.New(10), tree.New(10))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Same() = %v, want %v", err, context.Canceled)
	}
}
