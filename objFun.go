package main

const (
    JOBS_COUNT	  = iota // 0
    // JOBS_REJECTED
    ASSURANCE
	ENERGY
	SQUARED_CAPACITY
	MISUSED_RT
)

var weights = [6]float32{50., 45., -7.5, 5., -5}
var normalization_denom [5]float32

var objective_functions = []func(WorkerNode, WN_Status) float32{
	FO_jobsCount, FO_getAssurance, FO_energyCost, FO_squaredUnrequestedCapacity, FO_RTCapacityDoingNonRT,
	}

var objFunc_aggregator func([]float32) float32= weighted_sum_moo_method

const rt_space_extra_value float32 = 1. //4./3.

// Funzioni obiettivo
// Max Count Job
// Max Sum Assurance
// Min Active Nodes 		= Max -ActiveNodes
// Max (Free^2)
func FO_jobsCount(wn WorkerNode, status WN_Status) float32 {
	return float32(status.count)
}
func FO_jobsRejected(wn WorkerNode, status WN_Status) float32 {
	return normalization_denom[1] - float32(status.count)
}
func FO_getAssurance(wn WorkerNode, status WN_Status) float32 {
	return status.assurance
}
func FO_energyCost(wn WorkerNode, status WN_Status) float32 {
	if status.active {
		return float32(wn.Cost)
	}
	return 0
}
/*This actually gives more weight to free capacity in rt nodes than non rt
	the keepSign_centiSqr func, returns the squared value/100 in float32 but keeps the sign..
	therefore, a negative freeCpu will have a negative effect, as it should
*/
func FO_squaredFreeCapacity(wn WorkerNode, status WN_Status) float32 {
	var bonus float32 = 1
	if wn.RealTime {
		bonus = rt_space_extra_value
	}
	return keepSign_centiSqr(status.freeCPU) * bonus
}
func FO_squaredUnrequestedCapacity(wn WorkerNode, status WN_Status) float32 {
	var bonus float32 = 1
	if wn.RealTime {
		bonus = rt_space_extra_value
	}
	return keepSign_centiSqr(status.unrequestedCPU) * bonus
}
func FO_RTCapacityDoingNonRT(wn WorkerNode, status WN_Status) float32 {
	var ret float32 = 0
	if !wn.RealTime{
		return ret
	}
	// else implicito
	return float32(status.NO_Requested) //CPU C
}


// Aggregate FO
func weighted_sum_moo_method(fo_values []float32) float32 {
	var ret float32 = 0
	var normalized = normalize_fo_vector(fo_values, normalization_denom[:])
	/*Pre Process:
		- Number of pods accepted squared to give nonlinear penalty increment to not adding a pod
	*/
	// normalized[JOBS_COUNT] = (fo_values[JOBS_COUNT]*fo_values[JOBS_COUNT])/(normalization_denom[JOBS_COUNT]*normalization_denom[JOBS_COUNT])	
	// normalized[JOBS_REJECTED] = (fo_values[JOBS_REJECTED]*fo_values[JOBS_REJECTED])/(normalization_denom[JOBS_REJECTED]*normalization_denom[JOBS_REJECTED])	
	for o := range normalized {
		ret += normalized[o] * weights[o]
	}

	//Missed Pods Penalty
	//	fo_score - missedRatio %		missedRatio = 1-JobsCountRatio		
	//  fo_score - (1-JobsCountRatio)fo_score
	//  fo_score (1 - (1 - JobsCountRatio))
	//  fo_score * JobsCountRatio
	ret *= normalized[JOBS_COUNT]

	return ret
}

func normalize_fo_vector(fo_values []float32, denom []float32) []float32{
	var normalized []float32 = make([]float32, len(fo_values))
	for i := range normalized{
		normalized[i] = fo_values[i]/denom[i]
	}
	return normalized
}