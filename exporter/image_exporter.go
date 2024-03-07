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


type imageCollector struct {
	rawUsed				*prometheus.Desc
	writeStats			*prometheus.Desc
	readStats			*prometheus.Desc
	deleteStats			*prometheus.Desc

	vitastorConfig	*config.VitastorConfig
}


func newImageCollector(conf *config.VitastorConfig) *imageCollector {
	return &imageCollector{
		rawUsed: prometheus.NewDesc(prometheus.BuildFQName(namespace, "image", "raw_used"),
								"Image raw used in bytes",
								[]string{"pool_id", "image_num"},
								nil),
		writeStats: prometheus.NewDesc(prometheus.BuildFQName(namespace, "image", "write"),
								"Image write stat",
								[]string{"pool_id", "image_num", "stat_name"},
								nil),
		readStats: prometheus.NewDesc(prometheus.BuildFQName(namespace, "image", "read"),
								"Image read stat",
								[]string{"pool_id", "image_num", "stat_name"},
								nil),
		deleteStats: prometheus.NewDesc(prometheus.BuildFQName(namespace, "image", "delete"),
								"Image delete stat",
								[]string{"pool_id", "image_num", "stat_name"},
								nil),
		vitastorConfig: conf,
	}
}

func (collector *imageCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.rawUsed
	ch <- collector.writeStats
	ch <- collector.readStats
	ch <- collector.deleteStats
}

func (collector *imageCollector) Collect(ch chan<- prometheus.Metric) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   collector.vitastorConfig.VitastorEtcdUrls,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Error(err, "Unable to connect to etcd")
		return
	}
	defer cli.Close()

	//Collect pool ids
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

	for pool_id, _ := range pools {
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second * 20)
		imageStatsPath := collector.vitastorConfig.VitastorPrefix + "/inode/stats/" + pool_id
		imageStatsRaw, err := cli.Get(ctx2, imageStatsPath, clientv3.WithPrefix())
		cancel2()
		if err != nil {
			log.Error(err, "Unable to get image stats info")
			return
		}
		imageStats := make(map[string]config.VitastorImageStats)
		if imageStatsRaw.Count != 0 {
			for _, v := range imageStatsRaw.Kvs {
				var st config.VitastorImageStats
				err = json.Unmarshal(v.Value, &st)
				if err != nil {
					log.Error(err, "Unable to parse image stats")
				}
				image_num := strings.Split(string(v.Key),"/")[5]
				imageStats[image_num] = st
			}
		}

		for image, v := range imageStats {
			raw_used, err := v.RawUsed.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.rawUsed, prometheus.CounterValue, raw_used, pool_id, image)
			}
			read_count, err := v.ReadStats.Count.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_count, pool_id, image, "count")
			}
			read_usec, err := v.ReadStats.Usec.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_usec, pool_id, image, "usecs")
			}
			read_bytes, err := v.ReadStats.Bytes.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_bytes, pool_id, image, "bytes")
			}
			read_bps, err := v.ReadStats.Bps.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_bps, pool_id, image, "bps")
			}
			read_iops, err := v.ReadStats.Iops.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_iops, pool_id, image, "iops")
			}
			read_lat, err := v.ReadStats.Lat.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.readStats, prometheus.CounterValue, read_lat, pool_id, image, "lat")
			}

			write_count, err := v.WriteStats.Count.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_count, pool_id, image, "count")
			}
			write_usec, err := v.WriteStats.Usec.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_usec, pool_id, image, "usecs")
			}
			write_bytes, err := v.WriteStats.Bytes.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_bytes, pool_id, image, "bytes")
			}
			write_bps, err := v.WriteStats.Bps.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_bps, pool_id, image, "bps")
			}
			write_iops, err := v.WriteStats.Iops.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_iops, pool_id, image, "iops")
			}
			write_lat, err := v.WriteStats.Lat.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.writeStats, prometheus.CounterValue, write_lat, pool_id, image, "lat")
			}

			delete_count, err := v.DeleteStats.Count.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_count, pool_id, image, "count")
			}
			delete_usec, err := v.DeleteStats.Usec.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_usec, pool_id, image, "usecs")
			}
			delete_bytes, err := v.DeleteStats.Bytes.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_bytes, pool_id, image, "bytes")
			}
			delete_bps, err := v.DeleteStats.Bps.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_bps, pool_id, image, "bps")
			}
			delete_iops, err := v.DeleteStats.Iops.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_iops, pool_id, image, "iops")
			}
			delete_lat, err := v.DeleteStats.Lat.Float64()
			if err == nil {
				ch <- prometheus.MustNewConstMetric(collector.deleteStats, prometheus.CounterValue, delete_lat, pool_id, image, "lat")
			}
		}
	}
}
