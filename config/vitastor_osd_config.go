package config

type VitastorOSDState struct {
	Addresses []string `json:"addresses"`
	BlockstoreEnabled bool `json:"blockstore_enabled"`
	Host string `json:"host"`
	Port int `json:"port"`
	PrimaryEnabled bool `json:"primary_enabled"`
	State string `json:"up"`
}

type VitastorOSDStats struct {
	BlockstoreReady bool `json:"blockstore_ready"`
	DataBlockSize int `json:"data_block_size"`
	Host string `json:"host"`
	Free int `json:"free"`
	Size int `json:"size"`
	OpStats map[string]OSDStats `json:"op_stats"`
	SubopStats map[string]OSDStats `json:"subop_stats"`
	RecoveryStats map[string]OSDStats `json:"recovery_stats"`
}


type OSDStats struct {
	Bytes int `json:"bytes"`
	Count int `json:"count,omitempty"`
	Usecs int `json:"usecs,omitempty"`
}