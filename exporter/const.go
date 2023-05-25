package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/version"
	config "github.com/Antilles7227/vitastor-exporter/config"
)


const (
	namespace = "vitastor"
)

func Register(config *config.VitastorConfig) {
	poolCollector := newPoolCollector(config)
	monitorCollector := newMonitorCollector(config)
	osdCollector := newOsdCollector(config)
	prometheus.MustRegister(version.NewCollector("vitastor_exporter"))
	prometheus.MustRegister(poolCollector)
	prometheus.MustRegister(monitorCollector)
	prometheus.MustRegister(osdCollector)
	prometheus.Unregister(collectors.NewGoCollector())
}