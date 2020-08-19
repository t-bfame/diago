package scheduler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PodCollection maintains prometheus metric collectors
type PodCollection struct {
	totalCapacity   prometheus.Gauge
	currentCapacity prometheus.Gauge
	workerCount     prometheus.Gauge
}

func (pc *PodCollection) updateTotalCapacity(tot uint64) {
	pc.totalCapacity.Set(float64(tot))
}

func (pc *PodCollection) updateCurrentCapacity(cur uint64) {
	pc.currentCapacity.Set(float64(cur))
}

func (pc *PodCollection) updateWorkerCount(work uint64) {
	pc.workerCount.Set(float64(work))
}

// NewPodCollection returns a new prometheus metric collection
func NewPodCollection(labels map[string]string) *PodCollection {

	pc := PodCollection{
		totalCapacity: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_total_capacity",
			Help:        "Currently supported capacity",
			ConstLabels: prometheus.Labels(labels),
		}),
		currentCapacity: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_available_capacity",
			Help:        "Available capacity",
			ConstLabels: prometheus.Labels(labels),
		}),
		workerCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_worker_count",
			Help:        "Current workers in diago",
			ConstLabels: prometheus.Labels(labels),
		}),
	}

	return &pc
}
