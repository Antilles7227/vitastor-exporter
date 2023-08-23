package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "net/http/pprof"

	vconfig "github.com/Antilles7227/vitastor-exporter/config"
	exporter "github.com/Antilles7227/vitastor-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)


func main() {
	portArg := flag.Int("port", 8080, "Port to expose metrics. Default: 8080")
	uriArg := flag.String("metrics-path", "/metrics", "Path to expose metrics. Default: /metrics")
	vitastorConfArg := flag.String("vitastor-conf", "/etc/vitastor/vitastor.conf", "Path to vitastor.conf (to obtain etcd connection params). Default: /etc/vitastor/vitastor.conf")
	etcdUrlArg := flag.String("etcd-url", "", "Comma-separated list of etcd urls. WARNING: setting that param will override --vitastor-conf and ignore params in vitastor.conf. Default: empty")
	vitastorPrefix := flag.String("vitastor-prefix", "/vitastor", "Etcd tree prefix for Vitastor cluster info. Default: /vitastor")
	flag.Parse()

	config := vconfig.VitastorConfig{
		VitastorPrefix: *vitastorPrefix,
		VitastorEtcdUrls: strings.Split(*etcdUrlArg, ","),
	}
	log.Info("Trying to load vitastor.conf")
	err := loadConfiguration(*vitastorConfArg, &config)
	if err != nil {
		log.Info("Unable to load vitastor.conf, using command-line args")
	} else {
		log.Info("vitastor.conf loaded")
	}
	if *etcdUrlArg != "" {
		log.Info("etcdUrlArg is set, overriding params in vitastor.conf")
		config.VitastorEtcdUrls = strings.Split(*etcdUrlArg, ",")
		config.VitastorPrefix = *vitastorPrefix
	}
	
	exporter.Register(&config)

	http.Handle(*uriArg, promhttp.Handler())
    log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*portArg), nil))
}


func loadConfiguration(file string, config *vconfig.VitastorConfig) error {
	configFile, err := os.Open(file)
	if err != nil {
		log.Error(err, "Unable to open config")
		return err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return nil
}