package main

/* Worker Node */
type WN_Status struct {
	active 					bool
	count 					int
	
	//FreeCpu is the CPU capacity minus the sum of Pods' limits
	freeCPU					int
	//UnrequestedCpu is the CPU capacity minus the sum of Pods' requests
	unrequestedCPU			int
	assurance				float32

	NO_Requested			int
	NO_Limit				int
	LOW_Requested			int
	LOW_Limit				int
	HI_Requested			int
	HI_Limit				int
}	

	func (s WN_Status) clone() WN_Status{
		return WN_Status{
				active: s.active,
				count: s.count,
				freeCPU: s.freeCPU,
				unrequestedCPU: s.unrequestedCPU,
				assurance: s.assurance,
				NO_Requested: s.NO_Requested,
				NO_Limit: s.NO_Limit,
				LOW_Requested: s.LOW_Requested,
				LOW_Limit: s.LOW_Limit,
				HI_Requested: s.HI_Requested,
				HI_Limit: s.HI_Limit,
		}
	}