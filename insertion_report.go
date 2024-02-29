package main

import "fmt"

/* Worker Node */
type InsertionReport struct {
	pod							*Pod
	wn							[]WorkerNode
	n_nodes						int
	current_fo_components		[]float32
	current_fo_score			float32
	ifAdd_eligibility			[]bool
	ifAdd_fo_components			[][]float32
	ifAdd_fo_score				[]float32

	adviced_addIdx				int
}

	func new_InsertionReport(pod *Pod, wn []WorkerNode) InsertionReport{
		n_nodes := len(wn)
		return InsertionReport{
			pod: pod,
			wn : wn,
			ifAdd_eligibility: make([]bool, n_nodes),
			ifAdd_fo_components: make([][]float32, n_nodes),
			ifAdd_fo_score: make([]float32, n_nodes),
			adviced_addIdx: -1,
		}
	}

	func (ir *InsertionReport) log_current_fo(current_fo_components []float32, current_fo_score float32){
		if ir.current_fo_components == nil || len(ir.current_fo_components)!=len(current_fo_components){
			ir.current_fo_components = make([]float32, len(current_fo_components))
		}
		copy(ir.current_fo_components, current_fo_components)
		ir.current_fo_score = current_fo_score
	}

	func (ir *InsertionReport) log_node_ifAdd(node_idx int, eligible bool, ifAdd_fo_components []float32, ifAdd_fo_score float32){
		ir.ifAdd_eligibility[node_idx] = eligible
		ir.ifAdd_fo_score[node_idx] = ifAdd_fo_score
		if ir.ifAdd_fo_components[node_idx] == nil || len(ir.ifAdd_fo_components)!=len(ifAdd_fo_components){
			ir.ifAdd_fo_components[node_idx] = make([]float32, len(ifAdd_fo_components))
		}
		copy(ir.ifAdd_fo_components[node_idx], ifAdd_fo_components)

		// Is this best?
		//	This is better than currFO AND (There is no best, OR This is better than that)
		if ifAdd_fo_score > ir.current_fo_score && 
				( ir.adviced_addIdx < 0 || 
				  ifAdd_fo_score > ir.ifAdd_fo_score[ir.adviced_addIdx] ) {

			ir.adviced_addIdx = node_idx
		}
	}

	func (ir *InsertionReport) log_destNode(idx int){
		copy(ir.current_fo_components, ir.ifAdd_fo_components[idx])
		ir.current_fo_score = ir.ifAdd_fo_score[idx]
	}

	// Base
	func (ir InsertionReport) String() string {
		var ret string = "Current Situation \n"
		ret += fmt.Sprintf("\tPods accepted %.f / %.f \n", ir.current_fo_components[JOBS_COUNT], normalization_denom[JOBS_COUNT])
		// ret += fmt.Sprintf("\tPods rejected %.f / %.f \n", ir.current_fo_components[JOBS_REJECTED], normalization_denom[JOBS_REJECTED])
		ret += fmt.Sprintf("\tTotal Assurance %.f %% / %.f %%\n", ir.current_fo_components[ASSURANCE]*100, normalization_denom[ASSURANCE]*100)
		ret += fmt.Sprintf("\tEnergy Cost %.f / %.f \n", ir.current_fo_components[ENERGY], normalization_denom[ENERGY])
		ret += fmt.Sprintf("\tFree Capacity %0.2f / %.2f \n", ir.current_fo_components[SQUARED_CAPACITY], normalization_denom[SQUARED_CAPACITY])
		ret += fmt.Sprintf("\tMisused RT Capacity %0.2f / %.2f \n", ir.current_fo_components[MISUSED_RT], normalization_denom[MISUSED_RT])
		ret += fmt.Sprintf(">  Aggregate FO %.3f \n", ir.current_fo_score)

		var rt string
		for i := range ir.wn {
			if ir.wn[i].RealTime{ rt = "( RealTime )" } else { rt="" } 
			ret += fmt.Sprintf("\n\tWorker Node %d %s\n", ir.wn[i].ID, rt)
			if ir.ifAdd_eligibility[i]{
				ret += fmt.Sprintf("\t\tNode Capacity left %d (%d) / %d   ->  %d (%d) / %d\n", ir.wn[i].status.freeCPU, ir.wn[i].status.unrequestedCPU, ir.wn[i].CPU_Capacity, ir.wn[i].status.freeCPU-ir.pod.CPU_Limit, ir.wn[i].status.unrequestedCPU-ir.pod.CPU_Request, ir.wn[i].CPU_Capacity)
				ret += fmt.Sprintf("\t\tPods accepted %.f / %.f\n", ir.ifAdd_fo_components[i][JOBS_COUNT], normalization_denom[JOBS_COUNT])
				// ret += fmt.Sprintf("\t\tPods rejected %.f / %.f\n", ir.ifAdd_fo_components[i][JOBS_REJECTED], normalization_denom[JOBS_REJECTED])
				ret += fmt.Sprintf("\t\tTotal Assurance %.f %% / %.f %%\n", ir.ifAdd_fo_components[i][ASSURANCE]*100, normalization_denom[ASSURANCE]*100)
				ret += fmt.Sprintf("\t\tEnergy Cost %.f / %.f\n", ir.ifAdd_fo_components[i][ENERGY], normalization_denom[ENERGY])
				ret += fmt.Sprintf("\t\tFree Capacity %0.2f / %.2f\n", ir.ifAdd_fo_components[i][SQUARED_CAPACITY], normalization_denom[SQUARED_CAPACITY])
				ret += fmt.Sprintf("\t\tMisused RT Capacity %0.2f / %.2f \n", ir.ifAdd_fo_components[i][MISUSED_RT], normalization_denom[MISUSED_RT])
				ret += fmt.Sprintf("\t>  Aggregate FO %.3f\n", ir.ifAdd_fo_score[i])
			} else {
				ret += fmt.Sprintf("\t\tNode not eligible\n")
			}
		}
		return ret
	}

	func (ir InsertionReport) if_ith_str(i int) string {
		var ret string = ""
		if i==-1{
			ret += fmt.Sprintf("\tPods accepted %.f / %.f \n", ir.current_fo_components[JOBS_COUNT], normalization_denom[JOBS_COUNT])
			// ret += fmt.Sprintf("\tPods accepted %.f / %.f \n", ir.current_fo_components[JOBS_COUNT], normalization_denom[JOBS_REJECTED])
			ret += fmt.Sprintf("\tTotal Assurance %.f %% / %.f %%\n", ir.current_fo_components[ASSURANCE]*100, normalization_denom[ASSURANCE]*100)
			ret += fmt.Sprintf("\tEnergy Cost %.f / %.f \n", ir.current_fo_components[ENERGY], normalization_denom[ENERGY])
			ret += fmt.Sprintf("\tFree Capacity %0.2f / %.2f \n", ir.current_fo_components[SQUARED_CAPACITY], normalization_denom[SQUARED_CAPACITY])
			ret += fmt.Sprintf("\tMisused RT Capacity %0.2f / %.2f \n", ir.current_fo_components[MISUSED_RT], normalization_denom[MISUSED_RT])
			ret += fmt.Sprintf(">  Aggregate FO (same as before) %.3f \n", ir.current_fo_score)

		} else if ir.ifAdd_eligibility[i]{
			ret += fmt.Sprintf("\tPods accepted %.f / %.f\n", ir.ifAdd_fo_components[i][JOBS_COUNT], normalization_denom[JOBS_COUNT])
			// ret += fmt.Sprintf("\tPods Rejected %.f / %.f\n", ir.ifAdd_fo_components[i][JOBS_REJECTED], normalization_denom[JOBS_REJECTED])
			ret += fmt.Sprintf("\tTotal Assurance %.f %% / %.f %%\n", ir.ifAdd_fo_components[i][ASSURANCE]*100, normalization_denom[ASSURANCE]*100)
			ret += fmt.Sprintf("\tEnergy Cost %.f / %.f\n", ir.ifAdd_fo_components[i][ENERGY], normalization_denom[ENERGY])
			ret += fmt.Sprintf("\tFree Capacity %0.2f / %.2f\n", ir.ifAdd_fo_components[i][SQUARED_CAPACITY], normalization_denom[SQUARED_CAPACITY])
			ret += fmt.Sprintf("\tMisused RT Capacity %0.2f / %.2f \n", ir.ifAdd_fo_components[i][MISUSED_RT], normalization_denom[MISUSED_RT])
			ret += fmt.Sprintf("\t>  Aggregate FO %.3f\n", ir.ifAdd_fo_score[i])
		
		} else {
			ret += fmt.Sprintf("\t\tNode not eligible, you should not ask me this\n")
		}
		return ret
	}

	func (ir InsertionReport) onlyCurr() string {
		var ret string = ""
			ret += fmt.Sprintf("\tPods accepted %.f / %.f \n", ir.current_fo_components[JOBS_COUNT], normalization_denom[JOBS_COUNT])
			// ret += fmt.Sprintf("\tPods rejected %.f / %.f \n", ir.current_fo_components[JOBS_REJECTED], normalization_denom[JOBS_REJECTED])
			ret += fmt.Sprintf("\tTotal Assurance %.f %% / %.f %%\n", ir.current_fo_components[ASSURANCE]*100, normalization_denom[ASSURANCE]*100)
			ret += fmt.Sprintf("\tEnergy Cost %.f / %.f \n", ir.current_fo_components[ENERGY], normalization_denom[ENERGY])
			ret += fmt.Sprintf("\tFree Capacity %0.2f / %.2f \n", ir.current_fo_components[SQUARED_CAPACITY], normalization_denom[SQUARED_CAPACITY])
			ret += fmt.Sprintf("\tMisused RT Capacity %0.2f / %.2f \n", ir.current_fo_components[MISUSED_RT], normalization_denom[MISUSED_RT])
			ret += fmt.Sprintf(">  Aggregate FO: %.3f \n", ir.current_fo_score)
		return ret
	}

	func (ir InsertionReport) get_AdvicedAddIndex() int {
		return ir.adviced_addIdx
	}