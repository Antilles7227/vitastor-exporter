package exporter

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	config "github.com/Antilles7227/vitastor-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)


type osdCollector struct {
	params				*prometheus.Desc
	dataBlockSize 		*prometheus.Desc
	bitmapGranularity	*prometheus.Desc
	size 				*prometheus.Desc
	free 				*prometheus.Desc
	statsBytes 			*prometheus.Desc
	statsUsec			*prometheus.Desc
	statsCount			*prometheus.Desc

	vitastorConfig	*config.VitastorConfig
}


func newOsdCollector(conf *config.VitastorConfig) *osdCollector {
	return &osdCollector{
		params:		prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "status"),
								"OSD info. 1 if OSD up, 0 if down",
								[]string{"osd_num", "host", "port"}, 
								nil),
		dataBlockSize: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "data_block_size_bytes"),
								"OSD block size in bytes",
								[]string{"osd_num"},
								nil),
		bitmapGranularity: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "bitmap_granularity"),
								"OSD bitmap granularity in bytes",
								[]string{"osd_num"},
								nil),
		size: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "size_bytes"),
								"OSD size in bytes",
								[]string{"osd_num"},
								nil),
		free: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "free_bytes"),
								"OSD free size in bytes",
								[]string{"osd_num"},
								nil),
		statsBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "stat_bytes"),
								"OSD stat size",
								[]string{"osd_num", "stat_type", "stat_name"},
								nil),
		statsCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "stat_count"),
								"OSD stat count",
								[]string{"osd_num", "stat_type", "stat_name"},
								nil),
		statsUsec: prometheus.NewDesc(prometheus.BuildFQName(namespace, "osd", "stat_usec"),
								"OSD stat time in usecs",
								[]string{"osd_num", "stat_type", "stat_name"},
								nil),
		vitastorConfig: conf,
	}
}

func (collector *osdCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.params
	ch <- collector.dataBlockSize
	ch <- collector.bitmapGranularity
	ch <- collector.size
	ch <- collector.free
	ch <- collector.statsBytes
	ch <- collector.statsCount
	ch <- collector.statsUsec
}

func (collector *osdCollector) Collect(ch chan<- prometheus.Metric) {
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
	osdStatePath := collector.vitastorConfig.VitastorPrefix + "/osd/state"
	osdStateRaw, err := cli.Get(ctx, osdStatePath, clientv3.WithPrefix())
	cancel()
	if err != nil {
		log.Error(err, "Unable to get osd state info")
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second * 20)
	osdStatsPath := collector.vitastorConfig.VitastorPrefix + "/osd/stats"
	osdStatsRaw, err := cli.Get(ctx2, osdStatsPath, clientv3.WithPrefix())
	cancel2()
	if err != nil {
		log.Error(err, "Unable to get osd stats info")
		return
	}

	osdState := make(map[string]config.VitastorOSDState)
	osdStats := make(map[string]config.VitastorOSDStats)
	if osdStateRaw.Count != 0 {
		for _, v := range osdStateRaw.Kvs {
			var st config.VitastorOSDState
			err = json.Unmarshal(v.Value, &st)
			if err != nil {
				log.Error(err, "Unable to parse osd state")
			}
			osd_num := strings.Split(string(v.Key),"/")[4]
			osdState[osd_num] = st
		}
	}
	if osdStatsRaw.Count != 0 {
		for _, v := range osdStatsRaw.Kvs {
			var st config.VitastorOSDStats
			err = json.Unmarshal(v.Value, &st)
			if err != nil {
				log.Error(err, "Unable to parse osd stats")
			}
			osd_num := strings.Split(string(v.Key),"/")[4]
			osdStats[osd_num] = st
		}
	}

	for osd, v := range osdStats {
		if state, found := osdState[osd]; found	{
			ch <- prometheus.MustNewConstMetric(collector.params, prometheus.CounterValue, 1, osd, state.Host, strconv.Itoa(state.Port))
		} else {
			ch <- prometheus.MustNewConstMetric(collector.params, prometheus.CounterValue, 0, osd, v.Host, "unknown")
		}
		ch <- prometheus.MustNewConstMetric(collector.bitmapGranularity, prometheus.CounterValue, float64(v.BitmapGranularity), osd)
		ch <- prometheus.MustNewConstMetric(collector.dataBlockSize, prometheus.CounterValue, float64(v.DataBlockSize), osd)
		ch <- prometheus.MustNewConstMetric(collector.size, prometheus.CounterValue, float64(v.Size), osd)
		ch <- prometheus.MustNewConstMetric(collector.free, prometheus.CounterValue, float64(v.Free), osd)
		for op, stats := range v.OpStats {
			ch <- prometheus.MustNewConstMetric(collector.statsBytes, prometheus.CounterValue, float64(stats.Bytes), osd, "op", op)
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, float64(stats.Count), osd, "op", op)
			ch <- prometheus.MustNewConstMetric(collector.statsUsec, prometheus.CounterValue, float64(stats.Usec), osd, "op", op)
		}

		for subop, stats := range v.SubopStats {
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, float64(stats.Count), osd, "subop", subop)
			ch <- prometheus.MustNewConstMetric(collector.statsUsec, prometheus.CounterValue, float64(stats.Usec), osd, "subop", subop)
		}

		for rec, stats := range v.RecoveryStats {
			ch <- prometheus.MustNewConstMetric(collector.statsBytes, prometheus.CounterValue, float64(stats.Bytes), osd, "rec", rec)
			ch <- prometheus.MustNewConstMetric(collector.statsCount, prometheus.CounterValue, float64(stats.Count), osd, "rec", rec)
			ch <- prometheus.MustNewConstMetric(collector.statsUsec, prometheus.CounterValue, float64(stats.Usec), osd, "subop", rec)
		}
	}
}