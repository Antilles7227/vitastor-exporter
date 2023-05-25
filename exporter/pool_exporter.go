package exporter

import (
	"context"
	"encoding/json"
	"time"
	"strconv"
	config "github.com/Antilles7227/vitastor-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"

)


type poolCollector struct {
	params			*prometheus.Desc
	usedRawTb 		*prometheus.Desc
	totalRawTb 		*prometheus.Desc
	spaceEfficiency *prometheus.Desc
	rawToUsable 	*prometheus.Desc

	vitastorConfig	*config.VitastorConfig
}

func newPoolCollector(conf *config.VitastorConfig) *poolCollector {
	return &poolCollector{
		params:		prometheus.NewDesc(prometheus.BuildFQName(namespace, "pool", "info"),
								"Pool info",
								[]string{"pool_name", "pool_id", "pool_scheme", "pg_size", "parity_chunks", "pg_minsize", "pg_count", "failure_domain"}, 
								nil),
		usedRawTb: 	prometheus.NewDesc(prometheus.BuildFQName(namespace, "pool", "used_raw_tb"),
								"Raw used space of pool in TB",
								[]string{"pool_name", "pool_id"}, 
								nil),
		totalRawTb: prometheus.NewDesc(prometheus.BuildFQName(namespace, "pool", "total_raw_tb"),
								"Total raw space of pool in TB",
								[]string{"pool_name", "pool_id"},
								nil),
		spaceEfficiency: prometheus.NewDesc(prometheus.BuildFQName(namespace, "pool", "space_efficiency"),
								"Pool space usage efficiency",
								[]string{"pool_name", "pool_id"},
								nil),
		rawToUsable: prometheus.NewDesc(prometheus.BuildFQName(namespace, "pool", "raw_to_usable"),
								"Raw to usable space ratio",
								[]string{"pool_name", "pool_id"},
								nil),
		vitastorConfig: conf,
	}
}

func (collector *poolCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.params
	ch <- collector.usedRawTb
	ch <- collector.totalRawTb
	ch <- collector.rawToUsable
	ch <- collector.spaceEfficiency
}

//Collect implements required collect function for all promehteus collectors
func (collector *poolCollector) Collect(ch chan<- prometheus.Metric) {
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
	poolsPath := collector.vitastorConfig.VitastorPrefix + "/config/pools"
	poolsConfigRaw, err := cli.Get(ctx, poolsPath)
	cancel()
	if err != nil {
		log.Error(err, "Unable to retrive pools config")
		return
	}
	var pools map[string]config.VitastorPoolConfig
	if poolsConfigRaw.Count != 0 {
		err = json.Unmarshal(poolsConfigRaw.Kvs[0].Value, &pools)
		if err != nil {
			log.Error(err, "Unable to parse pools config block")
			return
		}
	} else {
		return
	}

	
	for id, v := range pools {
		poolStats := &config.VitastorPoolStats{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second * 20)
		poolStatsPath := collector.vitastorConfig.VitastorPrefix + "/pool/stats/" + id
		poolStatsRaw, err := cli.Get(ctx, poolStatsPath)
		cancel()
		if err != nil {
			log.Error(err, "Unable to retrive pool stats")
			return
		}
		if poolStatsRaw.Count != 0 {
			err = json.Unmarshal(poolStatsRaw.Kvs[0].Value, poolStats)
			if err != nil {
				log.Error(err, "Unable to parse pool stats")
			}
		}

		ch <- prometheus.MustNewConstMetric(collector.params, prometheus.GaugeValue, 1, v.Name, 
																						id,
																						v.Scheme,
																						strconv.Itoa(int(v.PGSize)),
																						strconv.Itoa(int(v.ParityChunks)),
																						strconv.Itoa(int(v.PGMinSize)),
																						strconv.Itoa(int(v.PGCount)),
																						v.FailureDomain)

		ch <- prometheus.MustNewConstMetric(collector.totalRawTb, prometheus.GaugeValue, poolStats.TotalRawTb, v.Name, id)
		ch <- prometheus.MustNewConstMetric(collector.usedRawTb, prometheus.GaugeValue, poolStats.UsedRawTb, v.Name, id)
		ch <- prometheus.MustNewConstMetric(collector.spaceEfficiency, prometheus.GaugeValue, poolStats.SpaceEfficiency, v.Name, id)
		ch <- prometheus.MustNewConstMetric(collector.rawToUsable, prometheus.GaugeValue, poolStats.RawToUsable, v.Name, id)
	}
}

