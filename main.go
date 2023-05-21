package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	vconfig "github.com/Antilles7227/vitastor-exporter/config"
)



var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
			Name: "myapp_processed_ops_total",
			Help: "The total number of processed events",
	})
)


func main() {
	portArg := flag.Int("port", 8080, "Port to expose metrics. Default: 8080")
	uriArg := flag.String("metrics-path", "/metrics", "Path to expose metrics. Default: /metrics")
	refreshIntArg := flag.Int("refresh-interval", 5, "Etcd check interval in seconds. Default: 5")
	vitastorConfArg := flag.String("vitastor-conf", "/etc/vitastor/vitastor.conf", "Path to vitastor.conf (to obtain etcd connection params). Default: /etc/vitastor/vitastor.conf")
	etcdUrlArg := flag.String("etcd-url", "", "Comma-separated list of etcd urls. WARNING: setting that param will override --vitastor-conf. Default: empty")
	vitastorPrefix := flag.String("vitastor-prefix", "/vitastor", "Etcd tree prefix for Vitastor cluster info. Default: /vitastor")

	config, err := loadConfiguration(*vitastorConfArg)
	if err != nil {
		if *etcdUrlArg != "" {
			log.Info("Unable to load vitastor.conf, using command-line args")
			config = vconfig.VitastorConfig{
				VitastorPrefix: *vitastorPrefix,
				VitastorEtcdUrls: strings.Split(*etcdUrlArg, ","),
			}
		} else {
		log.Error(err, "Unable to load vitastor.conf and unable to use")
		return
		}
	}

	if *etcdUrlArg != "" {
		log.Info("etcdUrlArg is set, overriding param in ")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.VitastorEtcdUrls,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err, "Unable to connect to etcd")
		return
	}
	defer cli.Close()

	go getMetrics(config, *refreshIntArg)

	http.Handle(*uriArg, promhttp.Handler())
    http.ListenAndServe(":" + string(int32(*portArg)), nil)
}

func getMetrics(config vconfig.VitastorConfig, refreshInterval int) {

}

func loadConfiguration(file string) (vconfig.VitastorConfig, error) {
	var config vconfig.VitastorConfig
	configFile, err := os.Open(file)
	if err != nil {
		log.Error(err, "Unable to open config")
		return vconfig.VitastorConfig{}, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}