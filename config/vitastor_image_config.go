package config

import "encoding/json"

type VitastorImageStats struct {
	RawUsed		json.Number	`json:"raw_used"`
	ReadStats	ImageStats	`json:"read"`
	WriteStats	ImageStats	`json:"write"`
	DeleteStats	ImageStats	`json:"delete"`
}

type ImageStats struct {
	Count json.Number `json:"count,omitempty"`
	Usec json.Number `json:"usec,omitempty"`
	Bytes json.Number `json:"bytes,omitempty"`
	Bps   json.Number `json:"bps,omitempty"`
	Iops  json.Number `json:"iops,omitempty"`
	Lat  json.Number `json:"lat,omitempty"`
}
