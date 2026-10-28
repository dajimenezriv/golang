package main

import (
	"fmt"
	"sync"
)

func main() {
	graph := map[int][]int{
		0: {1, 2},
		1: {2, 3},
		2: {3},
		3: {4},
		4: {},
	}
	queries := []int{0, 1, 2}
	numWorkers := 2

	results := ConcurrentBFSQueries(graph, queries, numWorkers)
	fmt.Println(results)
}

func ConcurrentBFSQueries(graph map[int][]int, queries []int, numWorkers int) map[int][]int {
	var (
		results   = make(map[int][]int)
		queriesCh = make(chan int)
		wg        sync.WaitGroup
	)

	for range numWorkers {
		wg.Go(func() {
			for q := range queriesCh {
				var (
					seen    = map[int]struct{}{q: {}}
					pending = []int{q}
					result  = []int{q}
				)

				for len(pending) > 0 {
					node := pending[0]
					pending = pending[1:]
					for _, neighbour := range graph[node] {
						if _, ok := seen[neighbour]; !ok {
							seen[neighbour] = struct{}{}
							pending = append(pending, neighbour)
							result = append(result, neighbour)
						}
					}
				}

				results[q] = result
			}
		})
	}

	for _, q := range queries {
		queriesCh <- q
	}
	close(queriesCh)
	wg.Wait()

	return results
}
