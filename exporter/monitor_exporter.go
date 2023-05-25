package exporter

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	config "github.com/Antilles7227/vitastor-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type monitorCollector struct {
	info			*prometheus.Desc

	vitastorConfig	*config.VitastorConfig
}

func newMonitorCollector(conf *config.VitastorConfig) *monitorCollector {
	return &monitorCollector{
		info:		prometheus.NewDesc(prometheus.BuildFQName(namespace, "monitor", "info"),
								"Monitor info, 1 is master, 0 is standby",
								[]string{"monitor_id", "monitor_hostname", "monitor_ip"}, 
								nil),
		vitastorConfig: conf,
	}
}

func (collector *monitorCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- collector.info
}

func (collector *monitorCollector) Collect(ch chan<- prometheus.Metric) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   collector.vitastorConfig.VitastorEtcdUrls,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err, "Unable to connect to etcd")
		return
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 20)
	masterMonPath := collector.vitastorConfig.VitastorPrefix + "/mon/master"
	masterMonRaw, err := cli.Get(ctx, masterMonPath)
	cancel()
	if err != nil {
		log.Error(err, "Unable to retrive master monitor block")
		return
	}
	var masterMonitor config.VitastorMonitor
	if masterMonRaw.Count != 0 {
		err = json.Unmarshal(masterMonRaw.Kvs[0].Value, &masterMonitor)
		if err != nil {
			log.Error(err, "Unable to parse master monitor block")
			return
		}
	} else {
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second * 20)
	monPath := collector.vitastorConfig.VitastorPrefix + "/mon/member"
	monRaw, err := cli.Get(ctx, monPath, clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Error(err, "Unable to retrive monitors list")
		return
	}
	monitors := make([]config.VitastorMonitor, monRaw.Count)
	if monRaw.Count != 0 {
		for i, v := range monRaw.Kvs {
			err = json.Unmarshal(v.Value, &monitors[i])
			if err != nil {
				log.Error(err, "Unable to parse pool stats")
			}
			id := strings.Split(string(v.Key), "/")[4]
			if id == masterMonitor.Id {
				ch <- prometheus.MustNewConstMetric(collector.info, prometheus.CounterValue, 1, string(v.Key), monitors[i].Hostname, monitors[i].Ip[0])
			} else {
				ch <- prometheus.MustNewConstMetric(collector.info, prometheus.CounterValue, 0, string(v.Key), monitors[i].Hostname, monitors[i].Ip[0])
			}
		}
	}
}