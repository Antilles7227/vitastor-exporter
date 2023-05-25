package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	vconfig "github.com/Antilles7227/vitastor-exporter/config"
	exporter "github.com/Antilles7227/vitastor-exporter/exporter"
)


func main() {
	portArg := flag.Int("port", 8080, "Port to expose metrics. Default: 8080")
	uriArg := flag.String("metrics-path", "/metrics", "Path to expose metrics. Default: /metrics")
	vitastorConfArg := flag.String("vitastor-conf", "/etc/vitastor/vitastor.conf", "Path to vitastor.conf (to obtain etcd connection params). Default: /etc/vitastor/vitastor.conf")
	etcdUrlArg := flag.String("etcd-url", "", "Comma-separated list of etcd urls. WARNING: setting that param will override --vitastor-conf. Default: empty")
	vitastorPrefix := flag.String("vitastor-prefix", "/vitastor", "Etcd tree prefix for Vitastor cluster info. Default: /vitastor")
	flag.Parse()
	
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
		log.Info("etcdUrlArg is set, overriding param in vitastor.conf")
		config.VitastorEtcdUrls = strings.Split(*etcdUrlArg, ",")
	}
	
	exporter.Register(&config)

	http.Handle(*uriArg, promhttp.Handler())
    log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*portArg), nil))
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