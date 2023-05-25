package config

type VitastorMonitor struct {
	Ip	[]string `json:"ip"`
	Hostname string `json:"hostname"`
	Id string `json:"id,omitempty"`
}