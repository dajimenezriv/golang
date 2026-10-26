package main

import (
	"fmt"
	"testing"

	"golang.org/x/tour/tree"
)

func TestWalk(t *testing.T) {
	n := 10
	ch := make(chan int)
	tr := tree.New(n)
	go Walk(tr, ch)
	for v := range ch {
		fmt.Println(v)
	}
}

func TestSame(t *testing.T) {
	x := Same(tree.New(10), tree.New(10))
	fmt.Println(x)
}
