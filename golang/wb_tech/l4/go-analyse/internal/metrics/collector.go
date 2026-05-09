package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Collector struct {
	startedAt time.Time
	gcPercent int
	now       func() time.Time

	processUptimeDesc       *prometheus.Desc
	gcConfiguredPercentDesc *prometheus.Desc
	gcLastRunAgeDesc        *prometheus.Desc
	gcLastRunTimestampDesc  *prometheus.Desc
}

func NewCollector(gcPercent int) *Collector {
	return &Collector{
		startedAt: time.Now(),
		gcPercent: gcPercent,
		now:       time.Now,
		processUptimeDesc: prometheus.NewDesc(
			"go_analyse_process_uptime_seconds",
			"Process uptime in seconds.",
			nil,
			nil,
		),
		gcConfiguredPercentDesc: prometheus.NewDesc(
			"go_analyse_gc_configured_percent",
			"Current GOGC target percent configured via debug.SetGCPercent.",
			nil,
			nil,
		),
		gcLastRunAgeDesc: prometheus.NewDesc(
			"go_analyse_gc_last_run_age_seconds",
			"Seconds since the last completed GC cycle, calculated from runtime.ReadMemStats.",
			nil,
			nil,
		),
		gcLastRunTimestampDesc: prometheus.NewDesc(
			"go_analyse_gc_last_run_timestamp_seconds",
			"Unix timestamp of the last completed GC cycle, calculated from runtime.ReadMemStats.",
			nil,
			nil,
		),
	}
}

func NewRegistry(gcPercent int) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		NewCollector(gcPercent),
	)

	return registry
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.processUptimeDesc
	ch <- c.gcConfiguredPercentDesc
	ch <- c.gcLastRunAgeDesc
	ch <- c.gcLastRunTimestampDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	now := c.now()

	ch <- prometheus.MustNewConstMetric(
		c.processUptimeDesc,
		prometheus.GaugeValue,
		now.Sub(c.startedAt).Seconds(),
	)
	ch <- prometheus.MustNewConstMetric(
		c.gcConfiguredPercentDesc,
		prometheus.GaugeValue,
		float64(c.gcPercent),
	)
	ch <- prometheus.MustNewConstMetric(
		c.gcLastRunAgeDesc,
		prometheus.GaugeValue,
		secondsSinceLastGC(stats.LastGC, now),
	)
	ch <- prometheus.MustNewConstMetric(
		c.gcLastRunTimestampDesc,
		prometheus.GaugeValue,
		float64(stats.LastGC)/1e9,
	)
}

func secondsSinceLastGC(lastGC uint64, now time.Time) float64 {
	if lastGC == 0 {
		return 0
	}

	return now.Sub(time.Unix(0, int64(lastGC))).Seconds()
}
