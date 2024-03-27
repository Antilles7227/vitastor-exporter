package exporter

import (
	"context"
	"encoding/json"
	"time"
	config "github.com/Antilles7227/vitastor-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)


type statsCollector struct {
	statsBytes 			*prometheus.Desc
	statsUsec			*prometheus.Desc
	statsCount			*prometheus.Desc
	statsBps			*prometheus.Desc
	statsLat			*prometheus.Desc
	statsIops			*prometheus.Desc
	objectBytes			*prometheus.Desc
	objectCount			*prometheus.Desc

	vitastorConfig	*config.VitastorConfig
}


func newStatsCollector(conf *config.VitastorConfig) *statsCollector {
	return &statsCollector{
		statsBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_bytes"),
								"Global stat size",
								[]string{"stat_type", "stat_name"},
								nil),
		statsCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_count"),
								"Global stat count",
								[]string{"stat_type", "stat_name"},
								nil),
		statsUsec: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_usec"),
								"Global stat time in usecs",
								[]string{"stat_type", "stat_name"},
								nil),
		statsBps: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_bps"),
								"Global stat bytes per second",
								[]string{"stat_type", "stat_name"},
								nil),
		statsLat: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_lat"),
								"Global stat latency in usecs",
								[]string{"stat_type", "stat_name"},
								nil),
		statsIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "stat_iops"),
								"Global stat IOPS",
								[]string{"stat_type", "stat_name"},
								nil),
		objectBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "object_bytes"),
								"Global object size in bytes",
								[]string{"object_type"},
								nil),
		objectCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "global", "object_count"),
								"Global object count",
								[]string{"object_type"},
								nil),
		vitastorConfig: conf,
	}
}


func (collector *statsCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.statsBytes
	ch <- collector.statsCount
	ch <- collector.statsUsec
	ch <- collector.statsBps
	ch <- collector.statsLat
	ch <- collector.statsIops
	ch <- collector.objectBytes
	ch <- collector.objectCount
}

func (collector *statsCollector) Collect(ch chan<- prometheus.Metric) {
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
	globalStatsPath := collector.vitastorConfig.VitastorPrefix + "/stats"
	globalStatsRaw, err := cli.Get(ctx, globalStatsPath)
	cancel()
	if err != nil {
		log.Error(err, "Unable to get global state info")
	}

	var globalStats config.VitastorStats
	if globalStatsRaw.Count != 0 {
		err = json.Unmarshal(globalStatsRaw.Kvs[0].Value, &globalStats)
		if err != nil {
			log.Error(err, "Unable to parse global stats")
		}
	} else {
		return
	}

	for op, stats := range globalStats.OpStats {
		bytes, err := stats.Bytes.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsBytes, prometheus.CounterValue, bytes, "op", op)
		}
		count, err := stats.Count.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, count, "op", op)
		}
		usecs, err := stats.Usec.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsUsec, prometheus.CounterValue, usecs, "op", op)
		}
		lat, err := stats.Lat.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsLat, prometheus.CounterValue, lat, "op", op)
		}
		bps, err := stats.Bps.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsBps, prometheus.CounterValue, bps, "op", op)
		}
		iops, err := stats.Iops.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsIops, prometheus.CounterValue, iops, "op", op)
		}
	}

	for subop, stats := range globalStats.SubopStats {
		count, err := stats.Count.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, count, "subop", subop)
		}
		usecs, err := stats.Usec.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsUsec, prometheus.CounterValue, usecs, "subop", subop)
		}
		lat, err := stats.Lat.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsLat, prometheus.CounterValue, lat, "subop", subop)
		}
		iops, err := stats.Iops.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsIops, prometheus.CounterValue, iops, "subop", subop)
		}
	}

	for rec, stats := range globalStats.RecoveryStats {
		bytes, err := stats.Bytes.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsBytes, prometheus.CounterValue, bytes, "rec", rec)
		}
		count, err := stats.Count.Float64()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, count, "rec", rec)
		}
	}

	clean, err := globalStats.ObjectCounts.Clean.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectCount, prometheus.CounterValue, clean, "clean")
	}
	degraded, err := globalStats.ObjectCounts.Degraded.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectCount, prometheus.CounterValue, degraded, "degraded")
	}
	incomplete, err := globalStats.ObjectCounts.Incomplete.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectCount, prometheus.CounterValue, incomplete, "incomplete")
	}
	misplaced, err := globalStats.ObjectCounts.Misplaced.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectCount, prometheus.CounterValue, misplaced, "misplaced")
	}
	object, err := globalStats.ObjectCounts.Object.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectCount, prometheus.CounterValue, object, "object")
	}


	bytes_clean, err := globalStats.ObjectBytes.Clean.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectBytes, prometheus.CounterValue, bytes_clean, "clean")
	}
	bytes_degraded, err := globalStats.ObjectBytes.Degraded.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectBytes, prometheus.CounterValue, bytes_degraded, "degraded")
	}
	bytes_incomplete, err := globalStats.ObjectBytes.Incomplete.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectBytes, prometheus.CounterValue, bytes_incomplete, "incomplete")
	}
	bytes_misplaced, err := globalStats.ObjectBytes.Misplaced.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectBytes, prometheus.CounterValue, bytes_misplaced, "misplaced")
	}
	bytes_object, err := globalStats.ObjectBytes.Object.Float64()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(collector.objectBytes, prometheus.CounterValue, bytes_object, "object")
	}
}