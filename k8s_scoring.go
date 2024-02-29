package main

func k8s_leastAllocated_score(node WorkerNode, p Pod) float32 {
	s := node.status
	return float32(s.NO_Requested + s.LOW_Requested + s.HI_Requested)
}

func k8s_mostAllocated_score(node WorkerNode, p Pod) float32 {
	return float32(node.status.unrequestedCPU)
}

func k8s_RequestedToCapacityRatio_score(node WorkerNode, p Pod) float32 {
	return float32(p.CPU_Request) / float32(node.CPU_Capacity)
}

func k8s_leastAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_mostAllocated_condition(score float32, best float32) bool {
	return score < best
}

func k8s_RequestedToCapacityRatio_condition(score float32, best float32) bool {
	return score > best
}
