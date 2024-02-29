package main

import "fmt"

func greedy_KP(pods []*Pod, assigned []bool, capacity int) ([]int, int, int){
	var minW int = -1
	var G []int
	var p int = 0
	for j := range pods{
		if !assigned[j]{
			if minW == -1 || minW > pods[j].CPU_Request{
				minW = pods[j].CPU_Request
			}
		}
	}
	if minW == -1 { return G, p, capacity }

	for j := range pods{
		if !assigned[j]{
			if capacity >= pods[j].CPU_Request{
				capacity -= pods[j].CPU_Request
				p ++
				G = append(G, j)
				assigned[j] = true

				if capacity < minW{ return G, p, capacity }
			}
		}
	}
	return G, p, capacity
}

func MK1(pods []*Pod, worker_nodes []WorkerNode) ([]bool, [][]int, []int, []int) {
	var assigned []bool = make([]bool, len(pods))
	var G [][]int = make([][]int, len(worker_nodes))
	var p, c []int = make([]int, len(worker_nodes)), make([]int, len(worker_nodes))
	for i := range worker_nodes{
		G[i], p[i], c[i] = greedy_KP(pods, assigned, worker_nodes[i].get_CapacityLeft())
		fmt.Println(assigned)
		fmt.Println(G)
	}
	return assigned, G, p, c
}