package config

type VitastorConfig struct {
	VitastorEtcdUrls []string `json:"etcd_address"`
	VitastorPrefix   string   `json:"etcd_prefix"`
}
