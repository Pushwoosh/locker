package locker

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	locksHeld     prometheus.Gauge
	locksAcquired prometheus.Counter
	locksReleased prometheus.Counter
	lockWait      prometheus.Counter
}

var metricsSource struct {
	locksHeld     *prometheus.GaugeVec
	locksAcquired *prometheus.CounterVec
	locksReleased *prometheus.CounterVec
	lockWait      *prometheus.CounterVec
}
var metricsSourceOnce sync.Once

func newMetrics(name string) *metrics {
	metricsSourceOnce.Do(func() {
		metricsSource.locksHeld = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "pw",
			Subsystem: "locker",
			Name:      "held_total",
			Help:      "Total amount of device locks currently held by the storage",
		}, []string{"name"})

		metricsSource.locksAcquired = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "pw",
			Subsystem: "locker",
			Name:      "acquired_total",
			Help:      "Total amount of device locks acquired by the storage",
		}, []string{"name"})

		metricsSource.locksReleased = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "pw",
			Subsystem: "locker",
			Name:      "released_total",
			Help:      "Total amount of device locks released by the storage",
		}, []string{"name"})

		metricsSource.lockWait = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "pw",
			Subsystem: "locker",
			Name:      "wait_seconds_total",
			Help:      "Total amount of seconds storage was inactive due to sleeping",
		}, []string{"name"})

		prometheus.MustRegister(metricsSource.locksHeld)
		prometheus.MustRegister(metricsSource.locksAcquired)
		prometheus.MustRegister(metricsSource.locksReleased)
		prometheus.MustRegister(metricsSource.lockWait)
	})

	return &metrics{
		locksHeld:     metricsSource.locksHeld.WithLabelValues(name),
		locksAcquired: metricsSource.locksAcquired.WithLabelValues(name),
		locksReleased: metricsSource.locksReleased.WithLabelValues(name),
		lockWait:      metricsSource.lockWait.WithLabelValues(name),
	}
}
