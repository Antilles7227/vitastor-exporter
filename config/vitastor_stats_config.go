package config

import "encoding/json"

type VitastorStats struct {
	OpStats       map[string]GlobalOpStats     `json:"op_stats"`
	SubopStats    map[string]GlobalOpStats     `json:"subop_stats"`
	RecoveryStats map[string]GlobalOpStats     `json:"recovery_stats"`
	ObjectCounts  GlobalObjectStats `json:"object_counts"`
	ObjectBytes   GlobalObjectStats `json:"object_bytes"`
}

type GlobalOpStats struct {
	Bytes json.Number `json:"bytes,omitempty"`
	Count json.Number `json:"count,omitempty"`
	Usec json.Number `json:"usec,omitempty"`
	Bps   json.Number `json:"bps,omitempty"`
	Iops  json.Number `json:"iops,omitempty"`
	Lat   json.Number `json:"lat,omitempty"`
}

type GlobalObjectStats struct {
	Object     json.Number `json:"object"`
	Clean      json.Number `json:"clean"`
	Misplaced  json.Number `json:"misplaced"`
	Degraded   json.Number `json:"degraded"`
	Incomplete json.Number `json:"incomplete"`
}
