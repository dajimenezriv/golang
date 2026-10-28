package main

import "fmt"

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
	results := make(map[int][]int)

	for _, q := range queries {
		var (
			seen    = make(map[int]struct{})
			result  = []int{}
			pending = []int{q}
		)
		for len(pending) > 0 {
			node := pending[0]
			pending = pending[1:]
			result = append(result, node)
			for _, neighbour := range graph[node] {
				if _, ok := seen[neighbour]; !ok {
					seen[neighbour] = struct{}{}
					pending = append(pending, neighbour)
				}
			}
		}
		results[q] = result
	}

	return results
}
