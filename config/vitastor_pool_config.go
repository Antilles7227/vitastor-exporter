package config

type VitastorPoolConfig struct {
	Name               string `json:"name"`
	Scheme             string `json:"scheme"`
	PGSize             int32  `json:"pg_size"`
	ParityChunks       int32  `json:"parity_chunks,omitempty"`
	PGMinSize          int32  `json:"pg_minsize"`
	PGCount            int32  `json:"pg_count"`
	FailureDomain      string `json:"failure_domain,omitempty"`
	MaxOSDCombinations int32  `json:"max_osd_combinations,omitempty"`
	BlockSize          int32  `json:"block_size,omitempty"`
	ImmediateCommit    string `json:"immediate_commit,omitempty"`
	OSDTags            interface{} `json:"osd_tags,omitempty"`
}

type VitastorPoolStats struct {
	UsedRawTb       float64 `json:"used_raw_tb"`
	TotalRawTb      float64 `json:"total_raw_tb"`
	RawToUsable     float64 `json:"raw_to_usable"`
	SpaceEfficiency float64 `json:"space_efficiency"`
}
