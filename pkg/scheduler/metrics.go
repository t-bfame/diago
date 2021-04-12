package scheduler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PodMetrics maintains prometheus metric collectors
type PodMetrics struct {
	totalCapacity   prometheus.Gauge
	currentCapacity prometheus.Gauge
	workerCount     prometheus.Gauge
}

func (pc *PodMetrics) updateCurrentCapacity(cur uint64) {
	pc.currentCapacity.Set(float64(cur))
}

func (pc *PodMetrics) cleanup() {
	prometheus.Unregister(pc.totalCapacity)
	prometheus.Unregister(pc.currentCapacity)
	prometheus.Unregister(pc.workerCount)
}

// NewPodMetrics returns a new prometheus metric collection
func NewPodMetrics(group string, instance InstanceID, totalCapacity uint64) *PodMetrics {

	labels := map[string]string{"worker_group": group, "worker_instance": string(instance)}

	pc := PodMetrics{
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

	pc.totalCapacity.Set(float64(totalCapacity))
	pc.currentCapacity.Set(float64(totalCapacity))
	pc.workerCount.Set(1)

	return &pc
}
