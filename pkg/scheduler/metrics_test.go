package scheduler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

const (
	workGroup           string     = "worker-1"
	instanceID1         InstanceID = "instance-1"
	capacity            uint64     = 10
	totalCapacityFmtStr            = `
		# HELP diago_total_capacity Currently supported capacity
		# TYPE diago_total_capacity gauge
		diago_total_capacity{ worker_group = "%s",worker_instance = "%s" } %f
	`
	currentCapacityFmtStr = `
		# HELP diago_available_capacity Available capacity
		# TYPE diago_available_capacity gauge
		diago_available_capacity{ worker_group = "%s",worker_instance = "%s" } %f
	`
	workerCountFmtStr = `
		# HELP diago_worker_count Current workers in diago
		# TYPE diago_worker_count gauge
		diago_worker_count{ worker_group = "%s",worker_instance = "%s" } %d
	`
)

func TestNewPodMetrics(t *testing.T) {
	expectedTotalCapacityStr := fmt.Sprintf(totalCapacityFmtStr, workGroup, instanceID1, float64(capacity))
	expectedCurrentCapacityStr := fmt.Sprintf(currentCapacityFmtStr, workGroup, instanceID1, float64(capacity))
	expectedWorkerCountStr := fmt.Sprintf(workerCountFmtStr, workGroup, instanceID1, 1)

	podMetrics := NewPodMetrics(workGroup, instanceID1, capacity)
	defer podMetrics.cleanup()
	if err := testutil.CollectAndCompare(podMetrics.totalCapacity, strings.NewReader(expectedTotalCapacityStr), "diago_total_capacity"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := testutil.CollectAndCompare(podMetrics.currentCapacity, strings.NewReader(expectedCurrentCapacityStr), "diago_available_capacity"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := testutil.CollectAndCompare(podMetrics.workerCount, strings.NewReader(expectedWorkerCountStr), "diago_worker_count"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestUpdateCurrentCapacity(t *testing.T) {
	expectedCurrentCapacity := uint64(5)

	podMetrics := NewPodMetrics(workGroup, instanceID1, capacity)
	defer podMetrics.cleanup()

	podMetrics.updateCurrentCapacity(expectedCurrentCapacity)
	if got := testutil.ToFloat64(podMetrics.currentCapacity); got != float64(expectedCurrentCapacity) {
		t.Errorf("want %f, got %f", float64(expectedCurrentCapacity), got)
	}
}
