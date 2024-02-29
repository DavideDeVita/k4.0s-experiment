package main

import "fmt"

const (
	CPU  = iota // 0
	Disk        // 1
	RAM         // 2
)

/* Worker Node */
type BasicResourceType struct {
	request int
	limit   int
}

func printable_wn_br(br [3]int, semisep string, sep string) string {
	return fmt.Sprintf("CPU:%s%d%sDisk:%s%d%sRAM:%s%d", semisep, br[CPU], sep, semisep, br[Disk], sep, semisep, br[RAM])
}

func printable_pod_br(br [3]BasicResourceType, semisep string, sep string) string {
	return fmt.Sprintf("CPU:%s%d (%d)%sDisk:%s%d (%d)%sRAM:%s%d (%d)", semisep, br[CPU].request, br[CPU].limit, sep, semisep, br[Disk].request, br[Disk].limit, sep, semisep, br[RAM].request, br[RAM].limit)
}