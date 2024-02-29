package main

import "fmt"

type Pod struct {
	ID          			int
	RealTime    			bool
	CPU_Request 			int
	CPU_Limit   			int
	Criticality 			int
	computation_left      	int
	completion_notified		bool
}

func (p Pod) _checkAssurance(assurance float32) bool {
	switch p.Criticality {
	case Low:
		return assurance >= theta_LO
	case High:
		return assurance >= theta_HI
	case No:
	default:
		return true
	}
	return true
}

func createRandomPod(id int, req_min int, req_max int, req_unit int, lim_max_ratio float32, expire_ratio float32) *Pod {
	var rnd = rand_01()
	var rt bool = rnd >= 0.45
	var c int = No
	if rt {
		rnd = rand_01()
		if rnd < 0.5 {
			c = Low
		} else {
			c = High
		}
	}else{
		rnd = rand_01()
		if rnd < 0.5 {
			c = No
		} else if rnd < 0.75{
			c = Low
		} else {
			c = High
		}
	}
	var req int = rand_ab_int(req_min, req_max)
	lim_max := int((float32(req) * lim_max_ratio))
	var lim int = rand_ab_int(req, lim_max)
	req *= req_unit
	lim *= req_unit
	cp_left := int(float32(req+lim)*(1+expire_ratio))

	return &Pod{
		ID:          id,
		RealTime:    rt,
		CPU_Request: req,
		CPU_Limit:   lim,
		Criticality: c,
		computation_left: cp_left,
		completion_notified : false,
	}
}

func (p *Pod) run(){
	if p.RealTime{
		p.computation_left -= rand_ab_int(p.CPU_Request, p.CPU_Limit)
	} else {
		p.computation_left -= rand_ab_int(0, p.CPU_Limit)
	}
}

func (p Pod) String() string {
	var crit string
	if p.Criticality == No {
		crit = "No"
	} else if (p.Criticality == Low) {
		crit = "Low"
	} else {
		crit = "High"
	}
	return fmt.Sprintf("Pod %d.\n\tReal time:\t\t%t\n\tCriticality\t\t%s\n\tCPU request:\t\t%d\tLimit:\t%d\n",
		p.ID, p.RealTime, crit, p.CPU_Request, p.CPU_Limit,
	)
}