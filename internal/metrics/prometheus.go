package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TODO: Add the all the other useful metrics like rate, tput etc

// PrometheusCollection maintains prometheus metric collectors
type PrometheusCollection struct {
	latencyMean prometheus.Gauge
	latencyMin  prometheus.Gauge
	latencyMax  prometheus.Gauge

	bytesIn  prometheus.Gauge
	bytesOut prometheus.Gauge

	requests prometheus.Gauge
	// rate     prometheus.Gauge

	success prometheus.Gauge
}

func (pc *PrometheusCollection) update(m *Metrics) {

	latencyMean := float64((*m).Latencies.Total) / float64((*m).Requests)
	latencyMin := float64((*m).Latencies.Min)
	latencyMax := float64((*m).Latencies.Max)

	pc.latencyMean.Set(latencyMean)
	pc.latencyMin.Set(latencyMin)
	pc.latencyMax.Set(latencyMax)

	bytesIn := float64((*m).BytesIn.Total) / float64((*m).Requests)
	bytesOut := float64((*m).BytesOut.Total) / float64((*m).Requests)

	pc.bytesIn.Set(bytesIn)
	pc.bytesOut.Set(bytesOut)

	requests := float64((*m).Requests)
	success := float64((*m).success) / float64((*m).Requests)

	pc.requests.Set(requests)
	pc.success.Set(success)
}

func (pc *PrometheusCollection) clear() {
	pc.latencyMean.Set(0)
	pc.latencyMin.Set(0)
	pc.latencyMax.Set(0)
	pc.bytesIn.Set(0)
	pc.bytesOut.Set(0)
	pc.requests.Set(0)
	pc.success.Set(0)
}

// NewPrometheusCollector returns a new prometheus metric collection
func NewPrometheusCollector(labels map[string]string) *PrometheusCollection {

	pc := PrometheusCollection{
		latencyMean: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_latency_mean",
			Help:        "Mean is the mean request latency",
			ConstLabels: prometheus.Labels(labels),
		}),
		latencyMin: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_latency_min",
			Help:        "Min is the minimum observed request latency",
			ConstLabels: prometheus.Labels(labels),
		}),
		latencyMax: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_latency_max",
			Help:        "Max is the maximum observed request latency",
			ConstLabels: prometheus.Labels(labels),
		}),
		bytesIn: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_bytes_in",
			Help:        "Bytes In is the computed incoming byte metrics",
			ConstLabels: prometheus.Labels(labels),
		}),
		bytesOut: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_bytes_out",
			Help:        "Bytes Out is the computed outgoing byte metrics",
			ConstLabels: prometheus.Labels(labels),
		}),
		requests: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_requests",
			Help:        "Requests is the total number of requests executed",
			ConstLabels: prometheus.Labels(labels),
		}),
		success: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "diago_success",
			Help:        "Success is the percentage of non-error responses",
			ConstLabels: prometheus.Labels(labels),
		}),
		// rate: promauto.NewGauge(prometheus.GaugeOpts{
		// 	Name:        "diago_rates",
		// 	Help:        "Rate is the rate of sent requests per second",
		// 	ConstLabels: prometheus.Labels(labels),
		// }),
	}

	return &pc
}
