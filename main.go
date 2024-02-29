package main

import (
	"fmt"
	"sort"
)

/* CONSTANTS */
const switch_cost float32 = 0
const allow_switch bool = true

var n_additional_nodes int = rand_ab_int(0, 2)

// const kill_a_node_at int = 200

/* GLOBAL VARIABLES */
// Number of Worker Nodes
var n int = rand_ab_int(4, 10)

// Number of Pods
var m int = rand_ab_int(300, 800)

// batch size for batch main
var batch_size int = 5

// Which algos am I comparing?
var testNames []string = []string{"Alastor", "K8s_LeastAllocated", "K8s_MostAllocated", "K8s_RequestedToCapacityRatio"}
var nTests int = len(testNames)

// Worker Nodes replicas for each algorithm
var workerNodes [][]WorkerNode = make([][]WorkerNode, nTests)

// List of all Pods
var allPods []*Pod

//var allPods []*Pod = make([]*Pod, m)

/*	*	*	*	*	*	Initialization	*	*	*	*	*	*/
func init() {
	for in := range testNames {
		workerNodes[in] = make([]WorkerNode, n)
	}

	/** Worker Nodes creation */
	for i := 0; i < n; i++ {
		wn := createRandomWorkerNode(
			i+1,         //Id
			100, 15, 40, // Capacity unit, min, max
			50, 15, 50, // Cost unit, min, max
			1, 4, // Asurance alpha beta and gamma min and max
		)
		fmt.Println(wn)
		// Every algo has the same nodes (replicas) inside
		for in := range testNames {
			workerNodes[in][i] = wn.copy()
		}
	}

	// Init the Normalization Vector
	{
		var totalFreeCapacity, totalActivation, rtCapacity float32 = 0, 0, 0
		var bonus float32 = 1
		for _, wn := range workerNodes[0] {
			if wn.RealTime {
				bonus = rt_space_extra_value
				rtCapacity += float32(wn.CPU_Capacity)
			} else {
				bonus = 1
			}
			totalFreeCapacity += keepSign_centiSqr(wn.CPU_Capacity) * bonus

			totalActivation += float32(wn.Cost)
		}
		normalization_denom = [5]float32{0, float32(n), totalActivation, totalFreeCapacity, rtCapacity}
	}
}

/*	*	*	*	*	*	Execution Modes	*	*	*	*	*	*/
func main() {
	const (
		Sequential = iota // 0
		Batch             // 1
		Test
	)

	var mode int = Sequential //Batch
	switch mode {
	case Sequential:
		main_sequential()
		break
	case Batch:
		main_batch()
		break
	case Test:
		mkp()
		break
	}
}

func mkp() {
	var pod *Pod
	var pods []*Pod = make([]*Pod, m)

	for j := 0; j < m; j++ {

		//Create Random Pod
		pod = createRandomPod(
			j,
			2, 15, 50, // CPU Required (min, max, unit)
			2.25,         // Cpu Limit Max Ratio: it will be random between required and Ratio*required
			rand_01()*20, // Cpu Limit Max Ratio: it will be random between required and Ratio*required
		)
		fmt.Println(pod)
		pods[j] = pod
	}
	batch_sort(pods, true, false)
	MK1(pods, workerNodes[0])
}

func main_sequential() {
	var pod *Pod
	var reports []InsertionReport
	var BenchmarkMatrix [][]float32 = make([][]float32, m)
	var AcceptanceRatioMatrix [][]float32 = make([][]float32, m)
	var AssuranceMatrix [][]float32 = make([][]float32, m)
	var ActvationCostMatrix [][]float32 = make([][]float32, m)
	var SquaredFreeSpaceMatrix [][]float32 = make([][]float32, m)
	var MisusedRTMatrix [][]float32 = make([][]float32, m)

	for j := 0; j < m; j++ {
		//Allocate BenchmarkMatrix row
		BenchmarkMatrix[j] = make([]float32, nTests+1)
		AcceptanceRatioMatrix[j] = make([]float32, nTests+1)
		AssuranceMatrix[j] = make([]float32, nTests+1)
		ActvationCostMatrix[j] = make([]float32, nTests+1)
		SquaredFreeSpaceMatrix[j] = make([]float32, nTests+1)
		MisusedRTMatrix[j] = make([]float32, nTests+1)

		//Create Random Pod
		pod = createRandomPod(
			j,
			3, 20, 50, // CPU Required (min, max, unit)
			2.25,         // Cpu Limit Max Ratio: it will be random between required and Ratio*required
			rand_01()*20, // Cpu Limit Max Ratio: it will be random between required and Ratio*required
		)

		//allPods[j] = pod
		// fmt.Printf("New Pod (%d - ID %d) %p\n", j, pod.ID, pod)

		reports = addPod_iteration(pod, true)

		// Benchmark
		BenchmarkMatrix[j][0] = float32(j)
		AcceptanceRatioMatrix[j][0] = float32(j)
		AssuranceMatrix[j][0] = float32(j)
		ActvationCostMatrix[j][0] = float32(j)
		SquaredFreeSpaceMatrix[j][0] = float32(j)
		MisusedRTMatrix[j][0] = float32(j)

		fmt.Println("# # # Benchmark # # #")
		for in := range testNames {
			fmt.Printf("%s\n%s\n\n", testNames[in], reports[in].onlyCurr())

			BenchmarkMatrix[j][in+1] = reports[in].current_fo_score
			AcceptanceRatioMatrix[j][in+1] = reports[in].current_fo_components[JOBS_COUNT] / normalization_denom[JOBS_COUNT]
			AssuranceMatrix[j][in+1] = reports[in].current_fo_components[ASSURANCE] / normalization_denom[ASSURANCE]
			ActvationCostMatrix[j][in+1] = reports[in].current_fo_components[ENERGY] / normalization_denom[ENERGY]
			SquaredFreeSpaceMatrix[j][in+1] = reports[in].current_fo_components[SQUARED_CAPACITY] / normalization_denom[SQUARED_CAPACITY]
			MisusedRTMatrix[j][in+1] = reports[in].current_fo_components[MISUSED_RT] / normalization_denom[MISUSED_RT]
		}

		/* * * 	In case of Alastor < max(K8s) print details	* * */
		// if reports[0].current_fo_score < max(reports[1].current_fo_score, reports[2].current_fo_score, reports[3].current_fo_score){
		// 	fmt.Println("# # # FAIL # # #")
		// 	for in := range(testNames){
		// 		fmt.Printf("# %s\n", testNames[in])
		// 		fmt.Println(reports[in])
		// 	}
		// }

		// Alastor - Every 3 Jobs, check wether to switch something
		if allow_switch && j%3 == 2 {
			alastor_curr_fo := reports[0].current_fo_score
			alastor_curr_fo_components := reports[0].current_fo_components
			from, what, to, ifSwitch_fo_score := check_wether_switch_a_pod(workerNodes[0], alastor_curr_fo, alastor_curr_fo_components)
			ifSwitch_fo_score += switch_cost
			if from != to {
				fmt.Printf("# Alastor # Advice: Pod %d could be moved from Worker Node %d to Worker Node %d to gain %.3f in fo score!\n\n",
					workerNodes[0][from].Pods[what].ID, workerNodes[0][from].ID, workerNodes[0][to].ID, ifSwitch_fo_score-alastor_curr_fo,
				)
				switched := workerNodes[0][from].removePod(what)
				workerNodes[0][to].addPod(switched)
			} else {
				fmt.Println("# Alastor # ----------- No good switch in mind!\n")
			}
		}

		// Run Pods and remove completed
		completed := runPods(allPods, j)

		if completed > 0 {
			normalization_denom[JOBS_COUNT] -= float32(completed)
			// normalization_denom[JOBS_REJECTED] -= float32(completed)
			for in := range testNames {
				// fmt.Printf("\n%s\n", testNames[in])
				for i := range workerNodes[in] {
					// fmt.Printf(" > Worker Node %d\n", workerNodes[in][i].ID)
					workerNodes[in][i].removeCompleted()
				}
			}
		}

		//Eventual new nodes
		if j > 0 && j%(1+m/(1+n_additional_nodes)) == 0 {
			//new wn
			fmt.Println("\nNew Worker Node added to cluster")
			wn := createRandomWorkerNode(
				n+1,         //Id
				100, 10, 30, // Capacity unit, min, max
				50, 15, 50, // Cost unit, min, max
				1, 4, // Asurance alpha beta and gamma min and max
			)
			fmt.Println(wn)
			//Add to all lists
			for in := range testNames {
				workerNodes[in] = append(workerNodes[in], wn.copy())
			}

			//Update denoms
			var freeCapacity float32
			if wn.RealTime {
				freeCapacity = rt_space_extra_value
				normalization_denom[MISUSED_RT] += float32(wn.CPU_Capacity)
			} else {
				freeCapacity = 1
			}
			freeCapacity *= keepSign_centiSqr(wn.CPU_Capacity)

			normalization_denom[ASSURANCE]++
			normalization_denom[ENERGY] += float32(wn.Cost)
			normalization_denom[SQUARED_CAPACITY] += freeCapacity
		}
	}
	matrixToCsv("benchmark.csv", BenchmarkMatrix[:], append([]string{"pod index"}, testNames[:]...), 2)
	matrixToCsv("acceptance.csv", AcceptanceRatioMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("assurance.csv", AssuranceMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("activationCost.csv", ActvationCostMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("squaredFreeSpace.csv", SquaredFreeSpaceMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("misusedRT.csv", MisusedRTMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	fmt.Printf("End of simulation:\n\tn: %d + %d\n\tm: %d\n\tweights: %.1f %.1f %.1f %.1f %.1f\n", n, n_additional_nodes, m, weights[0], weights[1], weights[2], weights[3], weights[4])
}

func main_batch() {
	var sort_asc bool = false //assuming p_i = 1 for each i
	var run_ratio float32 = 3. / 4.
	var pod_batch []*Pod = make([]*Pod, batch_size)
	var k8s_pod_batch []*Pod = make([]*Pod, batch_size)
	var reports []InsertionReport
	var BenchmarkMatrix [][]float32 = make([][]float32, m/batch_size)
	var AcceptanceRatioMatrix [][]float32 = make([][]float32, m)
	var AssuranceMatrix [][]float32 = make([][]float32, m)
	var ActvationCostMatrix [][]float32 = make([][]float32, m)
	var SquaredFreeSpaceMatrix [][]float32 = make([][]float32, m)
	var MisusedRTMatrix [][]float32 = make([][]float32, m)

	for j := 0; j < m/batch_size; j++ {
		for jj := 0; jj < batch_size; jj++ {
			pod_batch[jj] = createRandomPod(
				j*batch_size+jj,
				2, 15, 50, // CPU Required (min, max, unit)
				2.25,         // Cpu Limit Max Ratio: it will be random between required and Ratio*required
				rand_01()*20, // Cpu Limit Max Ratio: it will be random between required and Ratio*required
			)
			k8s_pod_batch[jj] = pod_batch[jj]
		}

		//Allocate BenchmarkMatrix row
		BenchmarkMatrix[j] = make([]float32, nTests+1)
		AcceptanceRatioMatrix[j] = make([]float32, nTests+1)
		AssuranceMatrix[j] = make([]float32, nTests+1)
		ActvationCostMatrix[j] = make([]float32, nTests+1)
		SquaredFreeSpaceMatrix[j] = make([]float32, nTests+1)
		MisusedRTMatrix[j] = make([]float32, nTests+1)

		fmt.Println("Batch")
		for i := range pod_batch {
			fmt.Printf("[%d]: Pod %d\tRT: %t\tCPU req: %d\n", i, pod_batch[i].ID, pod_batch[i].RealTime, pod_batch[i].CPU_Request)
		}
		alastor_batch_sort(pod_batch, sort_asc)
		batch_sort(k8s_pod_batch, sort_asc, false)

		fmt.Println("\nAlastor Batch")
		for i := range pod_batch {
			fmt.Printf("[%d]: Pod %d\tRT: %t\tCPU req: %d\n", i, pod_batch[i].ID, pod_batch[i].RealTime, pod_batch[i].CPU_Request)
		}
		fmt.Println("\nK8s Batch")
		for i := range k8s_pod_batch {
			fmt.Printf("[%d]: Pod %d\tRT: %t\tCPU req: %d\n", i, k8s_pod_batch[i].ID, k8s_pod_batch[i].RealTime, k8s_pod_batch[i].CPU_Request)
		}
		fmt.Println()

		for jj := 0; jj < batch_size; jj++ {
			reports = addPod_batchIteration(pod_batch[jj], k8s_pod_batch[jj], true)
		}

		// Benchmark
		BenchmarkMatrix[j][0] = float32(j)
		AcceptanceRatioMatrix[j][0] = float32(j)
		AssuranceMatrix[j][0] = float32(j)
		ActvationCostMatrix[j][0] = float32(j)
		SquaredFreeSpaceMatrix[j][0] = float32(j)
		MisusedRTMatrix[j][0] = float32(j)

		fmt.Println("# # # Benchmark # # #")
		for in := range testNames {
			fmt.Printf("%s\n%s\n\n", testNames[in], reports[in].onlyCurr())

			BenchmarkMatrix[j][in+1] = reports[in].current_fo_score
			AcceptanceRatioMatrix[j][in+1] = reports[in].current_fo_components[JOBS_COUNT] / normalization_denom[JOBS_COUNT]
			AssuranceMatrix[j][in+1] = reports[in].current_fo_components[ASSURANCE] / normalization_denom[ASSURANCE]
			ActvationCostMatrix[j][in+1] = reports[in].current_fo_components[ENERGY] / normalization_denom[ENERGY]
			SquaredFreeSpaceMatrix[j][in+1] = reports[in].current_fo_components[SQUARED_CAPACITY] / normalization_denom[SQUARED_CAPACITY]
			MisusedRTMatrix[j][in+1] = reports[in].current_fo_components[MISUSED_RT] / normalization_denom[MISUSED_RT]
		}

		/* * * 	In case of Alastor < max(K8s) print details	* * */
		// if reports[0].current_fo_score < max(reports[1].current_fo_score, reports[2].current_fo_score, reports[3].current_fo_score){
		// 	fmt.Println("# # # FAIL # # #")
		// 	for in := range(testNames){
		// 		fmt.Printf("# %s\n", testNames[in])
		// 		fmt.Println(reports[in])
		// 	}
		// }

		// Alastor - Every Batch of Jobs, check wether to switch something
		if allow_switch /*&& j%3 == 2*/ {
			alastor_curr_fo := reports[0].current_fo_score
			alastor_curr_fo_components := reports[0].current_fo_components
			from, what, to, ifSwitch_fo_score := check_wether_switch_a_pod(workerNodes[0], alastor_curr_fo, alastor_curr_fo_components)
			ifSwitch_fo_score += switch_cost
			if from != to {
				fmt.Printf("# Alastor # Advice: Pod %d could be moved from Worker Node %d to Worker Node %d to gain %.3f in fo score!\n\n",
					workerNodes[0][from].Pods[what].ID, workerNodes[0][from].ID, workerNodes[0][to].ID, ifSwitch_fo_score-alastor_curr_fo,
				)
				switched := workerNodes[0][from].removePod(what)
				workerNodes[0][to].addPod(switched)
			} else {
				fmt.Println("# Alastor # ----------- No good switch in mind!\n")
			}
		}

		// Run Pods (multiple times) and remove completed
		for jj := 0; jj < int(float32(batch_size)*run_ratio); jj++ {
			completed := runPods(allPods, j*batch_size)

			if completed > 0 {
				normalization_denom[JOBS_COUNT] -= float32(completed)
				// normalization_denom[JOBS_REJECTED] -= float32(completed)
				for in := range testNames {
					// fmt.Printf("\n%s\n", testNames[in])
					for i := range workerNodes[in] {
						//fmt.Printf(" > Worker Node %d\n", workerNodes[in][i].ID)
						workerNodes[in][i].removeCompleted()
					}
				}
			}
		}
	}
	matrixToCsv("benchmark.csv", BenchmarkMatrix[:], append([]string{"pod index"}, testNames[:]...), 2)
	matrixToCsv("acceptance.csv", AcceptanceRatioMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("assurance.csv", AssuranceMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("activationCost.csv", ActvationCostMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("squaredFreeSpace.csv", SquaredFreeSpaceMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	matrixToCsv("misusedRT.csv", MisusedRTMatrix[:], append([]string{"pod index"}, testNames[:]...), 3)
	fmt.Printf("End of simulation:\n\tn: %d + %d\n\tm: %d\n\tbatch size: %d\n\tweights: %.1f %.1f %.1f %.1f %.1f\n", n, n_additional_nodes, m, batch_size, weights[0], weights[1], weights[2], weights[3], weights[4])
}

/*	*	*	*	*	*	Alastor Switch	*	*	*	*	*	*/
func check_wether_switch_a_pod(workerNodes []WorkerNode, curr_fo_score float32, curr_fo_components []float32) (int, int, int, float32) {
	// From where and what do i remove
	var best_remove_fo_score float32
	var best_remove_fo_components []float32
	var best_pod_remove_index, best_wn_remove_index int = -1, -1

	// Fase di aggiustamento
	for wn_i, wn := range workerNodes {
		pod_index, fo_ifRemove, fo_components_ifRemove := wn.findMostTroublesomePod(
			curr_fo_score,
			curr_fo_components,
			objective_functions,
			objFunc_aggregator,
		)
		if pod_index != -1 && (best_wn_remove_index == -1 || fo_ifRemove > best_remove_fo_score) {
			best_remove_fo_score = fo_ifRemove
			best_remove_fo_components = fo_components_ifRemove
			best_pod_remove_index = pod_index
			best_wn_remove_index = wn_i
		}
	}
	// Is there a best to remove?
	if best_pod_remove_index != -1 {
		fmt.Println("# Alastor #  Considering switching from WN", best_wn_remove_index, "pod", best_pod_remove_index)
		var pod *Pod = workerNodes[best_wn_remove_index].Pods[best_pod_remove_index]
		// Where do I add ? (init to if same as remove)
		var best_wn_add_index int = best_wn_remove_index
		var best_add_fo_score float32 = curr_fo_score

		var eligible bool
		var ifStatus WN_Status
		var ifAdd_fo_components []float32 = make([]float32, len(objective_functions))
		var ifAdd_fo_score float32
		for i := range workerNodes {
			if i != best_wn_remove_index {
				wn := &workerNodes[i]
				wn.refreshStatus()
				eligible, ifStatus = wn.couldAddPod(*pod)

				if eligible {
					for o := range objective_functions {
						ifAdd_fo_components[o] = best_remove_fo_components[o] - objective_functions[o](*wn, *wn.status) + objective_functions[o](*wn, ifStatus)
					}
					ifAdd_fo_score = objFunc_aggregator(ifAdd_fo_components)
					fmt.Printf("\tIf added to WN %d, FO would become %.3f\n", i, ifAdd_fo_score)
					if ifAdd_fo_score > best_add_fo_score {
						best_add_fo_score = ifAdd_fo_score
						best_wn_add_index = i
					}
				}
			}
		}
		return best_wn_remove_index, best_pod_remove_index, best_wn_add_index, best_add_fo_score
	}
	return -1, -1, -1, 0
}

/*	*	*	*	*	*	Add Pod	*	*	*	*	*	*/
func addPod_iteration(pod *Pod, print_details bool) []InsertionReport {
	var addIdx []int = make([]int, nTests)
	var reports []InsertionReport = make([]InsertionReport, nTests)

	// on avg on n iters is O(1), implemented as ArrayLists internally
	allPods = append(allPods, pod)

	/* Normalization factors that change with every new pod
	# Pod count
	*/
	normalization_denom[JOBS_COUNT]++
	// normalization_denom[JOBS_REJECTED]++

	if print_details {
		fmt.Println("#########################################################")
		fmt.Println(pod)
	}

	// For each WorkerNode Vector, get the Add Pod Report
	for in := range testNames {
		reports[in] = get_addPodReport(
			workerNodes[in],
			*pod,
			objective_functions,
			objFunc_aggregator,
		)
	}

	//Alastor
	addIdx[0] = reports[0].get_AdvicedAddIndex()

	// K8S
	addIdx[1] = get_k8s_chooseWhereToInsert(
		workerNodes[1],
		*pod,
		k8s_leastAllocated_score,
		k8s_leastAllocated_condition,
	)
	addIdx[2] = get_k8s_chooseWhereToInsert(
		workerNodes[2],
		*pod,
		k8s_mostAllocated_score,
		k8s_mostAllocated_condition,
	)
	addIdx[3] = get_k8s_chooseWhereToInsert(
		workerNodes[3],
		*pod,
		k8s_RequestedToCapacityRatio_score,
		k8s_RequestedToCapacityRatio_condition,
	)

	// Adding
	for in := range testNames {
		idx := addIdx[in]
		if idx != -1 {
			if print_details {
				fmt.Printf("%s > Pod %d added to Worker Node %d\n\n", testNames[in], pod.ID, workerNodes[in][idx].ID)
			}
			workerNodes[in][idx].addPod(pod)
			reports[in].log_destNode(idx)
		} else if print_details {
			fmt.Printf("%s > Pod %d evaluated unsuitable for current configuration\n\n", testNames[in], pod.ID)
		}
	}

	return reports
}

func addPod_batchIteration(alastor_pod *Pod, k8s_pod *Pod, print_details bool) []InsertionReport {
	var addIdx []int = make([]int, nTests)
	var reports []InsertionReport = make([]InsertionReport, nTests)

	// on avg on n items is O(1), implemented as ArrayLists internally
	allPods = append(allPods, k8s_pod)

	/* Normalization factors that change with every new pod
	# Pod count
	*/
	normalization_denom[JOBS_COUNT]++
	// normalization_denom[JOBS_REJECTED]++

	if print_details {
		fmt.Println("#########################################################")
		fmt.Println("Alastor: ", alastor_pod)
		fmt.Println("K8s: ", k8s_pod)
	}

	// First iter add alastor pod, 2n+ on k82 pod
	var pod *Pod = alastor_pod
	for in := range testNames {
		reports[in] = get_addPodReport(
			workerNodes[in],
			*pod,
			objective_functions,
			objFunc_aggregator,
		)
		pod = k8s_pod
	}

	//Alastor
	addIdx[0] = reports[0].get_AdvicedAddIndex()

	// K8S
	addIdx[1] = get_k8s_chooseWhereToInsert(
		workerNodes[1],
		*k8s_pod,
		k8s_leastAllocated_score,
		k8s_leastAllocated_condition,
	)
	addIdx[2] = get_k8s_chooseWhereToInsert(
		workerNodes[2],
		*k8s_pod,
		k8s_mostAllocated_score,
		k8s_mostAllocated_condition,
	)
	addIdx[3] = get_k8s_chooseWhereToInsert(
		workerNodes[3],
		*k8s_pod,
		k8s_RequestedToCapacityRatio_score,
		k8s_RequestedToCapacityRatio_condition,
	)

	// Adding
	pod = alastor_pod
	for in := range testNames {
		idx := addIdx[in]
		if idx != -1 {
			if print_details {
				fmt.Printf("%s > Pod %d added to Worker Node %d\n\n", testNames[in], pod.ID, workerNodes[in][idx].ID)
			}
			workerNodes[in][idx].addPod(pod)
			reports[in].log_destNode(idx)
		} else if print_details {
			fmt.Printf("%s > Pod %d evaluated unsuitable for current configuration\n\n", testNames[in], pod.ID)
		}
		pod = k8s_pod
	}

	return reports
}

/*
This method assumes that each and every ObjFun is aggregative
Therefore, the global ObjFUn value is given as the sum of it components ObjFun
Es:	> 	Count Scheduled Jobs	// Maximize total number of Accepted Jobs

	>	Sum Assurance			// Maximize the average Assurance (Avg is directly proportional to the sum)
	>	- Activation Cost	 	// Minimize the energetic cost (inverted as Maximize the negative)
	>   Squared Free Space		// Maximize the presenze of Free space (squared to give more weight to )

Returns the current FO value vector of objective functions for node <i> if that node would be selected
*/
func get_addPodReport(
	workerNodes []WorkerNode,
	pod Pod,
	objFunc []func(WorkerNode, WN_Status) float32,
	fo_aggregator func([]float32) float32,
) InsertionReport {

	// Init the report
	var report InsertionReport = new_InsertionReport(&pod, workerNodes)

	// Evaluate the current status
	/*	At the end I have:
		eligible[]: is node 'i' eligible to add Pod?
		ifStatus[]: WN_Status of node 'i' if Pod were to be added there
		curr_ObjFun[]: Obj Function components vector of current setup (pre Pod p)
	*/
	var ifStatus []WN_Status = make([]WN_Status, len(workerNodes))
	var eligible []bool = make([]bool, len(workerNodes))
	var curr_ObjFun []float32 = make([]float32, len(objFunc))

	for i := range workerNodes {
		node := &workerNodes[i]

		node.refreshStatus()
		for o := range objFunc {
			curr_ObjFun[o] += objFunc[o](*node, *node.status)
		}

		eligible[i], ifStatus[i] = node.couldAddPod(pod)
	}

	report.log_current_fo(curr_ObjFun, fo_aggregator(curr_ObjFun))

	// Computing Obj Function Vector If Node <i> is selected (for each i)
	var ifObjFun []float32 = make([]float32, len(objFunc))

	for i := range workerNodes {
		node := &workerNodes[i]

		if !eligible[i] {
			report.log_node_ifAdd(i, false, nil, -1)
		} else {
			for o := range objFunc {
				/* Every FO is aggregative..
				therefore it is possible to remove from each component's current setup
				this node's component and add the one evaluated with the if status
				*/
				ifObjFun[o] = curr_ObjFun[o] - objFunc[o](*node, *node.status) + objFunc[o](*node, ifStatus[i])
			}
			report.log_node_ifAdd(i, true, ifObjFun[:], fo_aggregator(ifObjFun))
		}
	}
	return report
}

func get_k8s_chooseWhereToInsert(
	workerNodes []WorkerNode,
	pod Pod,
	k8s_scoringFun func(WorkerNode, Pod) float32,
	k8s_condition func(float32, float32) bool,
) int {

	var eligible bool
	var score, best float32
	var argbest int = -1

	for i := range workerNodes {
		node := &workerNodes[i]

		node.refreshStatus()
		eligible, _ = node.couldAddPod(pod)

		if eligible {
			score = k8s_scoringFun(*node, pod)
			if argbest == -1 || k8s_condition(score, best) {
				best = score
				argbest = i
			}
		}
	}
	return argbest
}

/*	*	*	*	*	*	Others	*	*	*	*	*	*/
// Run pods and reduce computation left
func runPods(allPods []*Pod, lim int) int {
	var completed int = 0
	for i := range allPods {
		if i == lim {
			return completed
		}
		var p *Pod = allPods[i]
		p.run()
		if !p.completion_notified && p.computation_left <= 0 {
			p.completion_notified = true
			completed++
			fmt.Printf("pod %d (%p) completed\n", p.ID, p)
		}
	}
	return completed
}

func batch_sort(pods []*Pod, ascending bool, withBonus bool) {
	// Define a custom less function
	lessFunc := func(ascending bool, withBonus bool) func(i, j int) bool {
		return func(i, j int) bool {
			// Sort desc by Req Size (with RT bonus)
			var i_bonus, j_bonus float32 = 1., 1.
			if withBonus {
				if pods[i].RealTime {
					i_bonus = rt_space_extra_value
				}
				if pods[j].RealTime {
					j_bonus = rt_space_extra_value
				}
			}
			if ascending {
				return (float32(pods[i].CPU_Request) * i_bonus) < (float32(pods[j].CPU_Request) * j_bonus)
			} else {
				return (float32(pods[i].CPU_Request) * i_bonus) > (float32(pods[j].CPU_Request) * j_bonus)
			}
		}
	}

	// Use sort.Slice to sort the slice using the custom less function
	sort.Slice(pods, lessFunc(ascending, withBonus))
}

func alastor_batch_sort(pods []*Pod, ascending bool) {
	/* For Alastor Algo, we insert first any RT Pod and then the others, to ensure that no RT space is filled by non rt */
	lessFunc := func(ascending bool) func(i, j int) bool {
		return func(i, j int) bool {
			var rt_i, rt_j bool = pods[i].RealTime, pods[j].RealTime
			if rt_i && !rt_j {
				return true
			} else if !rt_i && rt_j {
				return false
			}
			// Sort desc by Req Size
			if ascending {
				return float32(pods[i].CPU_Request) < float32(pods[j].CPU_Request)
			} else {
				return float32(pods[i].CPU_Request) > float32(pods[j].CPU_Request)
			}
		}
	}

	// Use sort.Slice to sort the slice using the custom less function
	sort.Slice(pods, lessFunc(ascending))
}
