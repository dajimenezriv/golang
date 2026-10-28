package main

import (
	"fmt"
	"slices"
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
	if numWorkers <= 0 {
		return map[int][]int{}
	}

	var (
		mu        sync.Mutex
		results   = make(map[int][]int)
		queriesCh = make(chan int)
		wg        sync.WaitGroup
	)

	for range numWorkers {
		wg.Go(func() {
			seen := make(map[int]struct{})
			var result []int

			for q := range queriesCh {
				clear(seen)
				seen[q] = struct{}{}
				result = append(result[:0], q)

				for head := 0; head < len(result); head++ {
					for _, neighbour := range graph[result[head]] {
						if _, ok := seen[neighbour]; !ok {
							seen[neighbour] = struct{}{}
							result = append(result, neighbour)
						}
					}
				}

				mu.Lock()
				results[q] = slices.Clone(result)
				mu.Unlock()
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
