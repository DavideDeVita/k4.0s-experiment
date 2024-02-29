package main

import (
	"fmt"
)

const Alpha_max = 5
const Beta_max = 5
const Gamma_max = 10
const Assurance_Denom = Alpha_max + Beta_max + Gamma_max

/* Worker Node */
type WorkerNode struct {
	ID           int
	RealTime     bool
	CPU_Capacity int
	Cost         int
	alpha        float32
	beta         float32
	Pods         []*Pod
	status       *WN_Status
}

// Base
func createRandomWorkerNode(id int, capacity_unit int, capacity_min int, capacity_max int, cost_unit int, cost_min int, cost_max int, Ass_min int, ass_max int) WorkerNode {
	var rt = rand_01() >= 0.5
	var capacity int = rand_ab_int(capacity_min, capacity_max) * capacity_unit
	var cost int = rand_ab_int(cost_min, cost_max) * cost_unit
	var alpha = rand_ab_int(Ass_min, ass_max)
	var beta = rand_ab_int(Ass_min, ass_max)
	return WorkerNode{
		ID:           id,
		RealTime:     rt,
		CPU_Capacity: capacity,
		Cost:         cost,
		alpha:        float32(alpha),
		beta:         float32(beta),
		Pods:         []*Pod{},
	}
}

// Base
func (wn WorkerNode) copy() WorkerNode {
	return WorkerNode{
		ID:           wn.ID,
		RealTime:     wn.RealTime,
		CPU_Capacity: wn.CPU_Capacity,
		Cost:         wn.Cost,
		alpha:        wn.alpha,
		beta:         wn.beta,
		Pods:         []*Pod{},
	}
}

func (wn WorkerNode) String() string {
	return fmt.Sprintf("Worker Node %d (%d Pods currenty deployed).\n\tReal time:\t\t%t\n\tCPU capacity:\t\t%d\n\tActivation cost\t\t%d\n\tAssurance (alpha, beta, gammaMax)\t%f\t%f\n",
		wn.ID, len(wn.Pods), wn.RealTime, wn.CPU_Capacity, wn.Cost, wn.alpha, wn.beta,
	)
}

// Getters
func (wn WorkerNode) getCPU_Usages() (int, int, int, int, int, int) {
	var HI_Limit, HI_Requested, LOW_Limit, LOW_Requested, NO_Limit, NO_Requested int
	for _, pod := range wn.Pods {
		switch pod.Criticality {
		case No:
			NO_Limit += pod.CPU_Limit
			NO_Requested += pod.CPU_Request
			break
		case Low:
			LOW_Limit += pod.CPU_Limit
			LOW_Requested += pod.CPU_Request
			break
		case High:
			HI_Limit += pod.CPU_Limit
			HI_Requested += pod.CPU_Request
			break
		}
	}
	return NO_Requested, NO_Limit, LOW_Requested, LOW_Limit, HI_Requested, HI_Limit
}

func (wn WorkerNode) get_CapacityLeft() int {
	var cl int = wn.CPU_Capacity
	for _, p := range wn.Pods {
		cl -= p.CPU_Request
	}
	return cl
}

func (wn WorkerNode) getAssurance(NO_Requested, NO_Limit, LOW_Requested, LOW_Limit, HI_Requested, HI_Limit int) float32 {
	var gamma float32 = 0

	// Criterio di Garantibilità dei Job Altamente Critici
	if HI_Limit <= wn.CPU_Capacity {
		gamma += 0.2
		// Criterio di Sostenibilità dei Job Critici
		if LOW_Requested+HI_Limit <= wn.CPU_Capacity {
			gamma += 0.2
			// Criterio di Garantibilità dei Job Critici
			if LOW_Limit+HI_Limit <= wn.CPU_Capacity {
				gamma += 0.1
			}
			// Criterio di Sostenibilità dei Job Critici e Non Interruzione
			if NO_Requested+LOW_Requested+HI_Limit <= wn.CPU_Capacity {
				gamma += 0.1
			}
			// Criterio di Non Interruzione
			if gamma > 0.4 && (NO_Requested+LOW_Limit+HI_Limit <= wn.CPU_Capacity) {
				gamma += 0.2
				// Criterio di Assoluta gestibilità
				if NO_Limit+LOW_Limit+HI_Limit <= wn.CPU_Capacity {
					gamma += 0.2
				}
			}
		}
	}
	gamma *= Gamma_max
	return (wn.alpha + wn.beta + gamma) / Assurance_Denom
}

func (wn WorkerNode) getStatus() *WN_Status {
	return wn.status
}

func (wn *WorkerNode) refreshStatus() {
	if wn.status == nil {
		wn.status = &WN_Status{}
	}
	/* CPU */
	wn.status.active = len(wn.Pods) > 0
	wn.status.count = len(wn.Pods)
	var NO_Requested, NO_Limit, LOW_Requested, LOW_Limit, HI_Requested, HI_Limit int = wn.getCPU_Usages()

	wn.status.freeCPU = wn.CPU_Capacity - (NO_Limit + LOW_Limit + HI_Limit)
	wn.status.unrequestedCPU = wn.CPU_Capacity - (NO_Requested + LOW_Requested + HI_Requested)

	wn.status.NO_Requested = NO_Requested
	wn.status.NO_Limit = NO_Limit
	wn.status.LOW_Requested = LOW_Requested
	wn.status.LOW_Limit = LOW_Limit
	wn.status.HI_Requested = HI_Requested
	wn.status.HI_Limit = HI_Limit

	wn.status.assurance = wn.getAssurance(NO_Requested, NO_Limit, LOW_Requested, LOW_Limit, HI_Requested, HI_Limit)
}

// Bool Match
func (wn WorkerNode) baseRequirementsMatch(p Pod) bool {
	return (wn.RealTime || !p.RealTime) && wn.get_CapacityLeft() >= p.CPU_Request
}
func (wn WorkerNode) advancedRequirementsMatch(p Pod) bool {
	return true
}

// Pod Insertion
func (wn *WorkerNode) addPod(p *Pod) {
	wn.Pods = append(wn.Pods, p)
}

func (wn WorkerNode) couldAddPod(p Pod) (bool, WN_Status) {
	var es = wn.status.clone()
	var eligible bool = (wn.RealTime || !p.RealTime) && wn.advancedRequirementsMatch(p)
	if eligible {
		// If i deploy a Pod it will always become active, regardless of its previous state
		es.active = true
		es.count += 1

		// Updating values
		es.freeCPU -= p.CPU_Limit
		es.unrequestedCPU -= p.CPU_Request

		switch p.Criticality {
		case No:
			es.NO_Requested += p.CPU_Request
			es.NO_Limit += p.CPU_Limit
			break
		case Low:
			es.LOW_Requested += p.CPU_Request
			es.LOW_Limit += p.CPU_Limit
			break
		case High:
			es.HI_Requested += p.CPU_Request
			es.HI_Limit += p.CPU_Limit
			break
		}

		es.assurance = wn.getAssurance(
			es.NO_Requested,
			es.NO_Limit,
			es.LOW_Requested,
			es.LOW_Limit,
			es.HI_Requested,
			es.HI_Limit)

		// Base (Kubernetes) Eligibility
		eligible = es.unrequestedCPU >= 0 &&
			p._checkAssurance(es.assurance)

		// Early exit
		if !eligible {
			return false, es
		}

		// Does every critical pod have the minimum requested assurance?
		for _, pod := range wn.Pods {
			eligible = pod._checkAssurance(es.assurance)

			// Early exit
			if !eligible {
				return false, es
			}
		}
	}
	return eligible, es
}

// Evt Switch
func (wn *WorkerNode) findMostTroublesomePod(
	global_fo float32,
	global_fo_components []float32,
	objFunc []func(WorkerNode, WN_Status) float32,
	fo_aggregator func([]float32) float32,
) (int, float32, []float32) {

	if len(wn.Pods) == 0 {
		return -1, 0., nil
	}
	wn.refreshStatus()

	//Current Contribute fo this node
	var curr_fo_components []float32 = make([]float32, len(objFunc))
	for o := range objFunc {
		curr_fo_components[o] += objFunc[o](*wn, *wn.status)
	}

	// If remove
	var fo_components_ifRemove []float32 = make([]float32, len(objFunc))
	var fo_score_ifRemove float32
	var argmax int = -1
	var best_gain float32
	var best_gain_components []float32
	for i, pod := range wn.Pods {
		if_st := wn.status.clone()

		// How would it be if <pod> was not here
		if_st.count -= 1
		if_st.active = if_st.count > 0

		if_st.freeCPU += pod.CPU_Limit
		if_st.unrequestedCPU += pod.CPU_Request

		switch pod.Criticality {
		case No:
			if_st.NO_Requested -= pod.CPU_Request
			if_st.NO_Limit -= pod.CPU_Limit
			break
		case Low:
			if_st.LOW_Requested -= pod.CPU_Request
			if_st.LOW_Limit -= pod.CPU_Limit
			break
		case High:
			if_st.HI_Requested -= pod.CPU_Request
			if_st.HI_Limit -= pod.CPU_Limit
			break
		}

		if_st.assurance = wn.getAssurance(
			if_st.NO_Requested,
			if_st.NO_Limit,
			if_st.LOW_Requested,
			if_st.LOW_Limit,
			if_st.HI_Requested,
			if_st.HI_Limit,
		)

		// FO
		for o := range objFunc {
			fo_components_ifRemove[o] = global_fo_components[o] - curr_fo_components[o] + objective_functions[o](*wn, if_st)
		}
		fo_score_ifRemove = fo_aggregator(fo_components_ifRemove)

		// Favour removing non-RT Pods from RT Worker Nodes
		if wn.RealTime && !pod.RealTime {
			fo_score_ifRemove += abs(fo_score_ifRemove) * 0.15
		}

		// in-loop check max
		if argmax == -1 || fo_score_ifRemove > best_gain {
			argmax = i
			best_gain = fo_score_ifRemove
			best_gain_components = fo_components_ifRemove
		}
	}
	if argmax != -1 {
		fmt.Printf("# Alastor # Worker Node %d, would benefit %f from removing Pod %d\n", wn.ID, best_gain-global_fo, wn.Pods[argmax].ID)
	}
	return argmax, best_gain, best_gain_components
}

// Pod Removal
func (wn *WorkerNode) removePod(idx int) *Pod {
	p := wn.Pods[idx]
	wn.Pods = append(wn.Pods[:idx], wn.Pods[idx+1:]...)
	return p
}

func (wn *WorkerNode) removeCompleted() {
	var toRemove []int = []int{}
	for i := len(wn.Pods) - 1; i > -1; i-- {
		var p *Pod = wn.Pods[i]
		// fmt.Printf("Pod %d (%p) c left %d\n", p.ID, p, p.computation_left)
		if p.computation_left <= 0 {
			toRemove = append(toRemove, i)
		}
	}
	for _, r_idx := range toRemove {
		// fmt.Printf("\tPod %d completed n removed\n", wn.Pods[r_idx].ID)
		wn.removePod(r_idx)
	}
}
