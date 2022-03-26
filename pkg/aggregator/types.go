package aggregator

import (
	"errors"
	"log"
	"time"

	pytypes "github.com/golang/protobuf/ptypes"
	m "github.com/t-bfame/diago/pkg/model"
	aggregator "github.com/t-bfame/diago/proto-gen/aggregator"
)

// InstanceID Pod instance ID for identification
type InstanceID string

// Event for internal communication
type Event interface {
	getJobID() m.JobID
}

// Incoming messages from aggregators to leader
type Incoming interface {
	getJobID() m.JobID
}


// AggMetrics message
type AggMetrics struct {
	TestId string
	InstanceId string
	JobId string
	// Latencies holds computed request latency metrics.
	Latencies []*durationpb.Duration `json:"latencies"`

	// Histogram, only if requested
	// Histogram *Histogram `json:"buckets,omitempty"`

	// BytesIn holds computed incoming byte metrics.
	BytesIn uint64 `json:"bytes_in"`

	// BytesOut holds computed outgoing byte metrics.
	BytesOut uint64 `json:"bytes_out"`

	// Earliest is the earliest timestamp in a Result set.
	Earliest time.Time `json:"earliest"`

	// Latest is the latest timestamp in a Result set.
	Latest time.Time `json:"latest"`

	// End is the latest timestamp in a Result set plus its latency.
	End time.Time `json:"end"`

	// Requests is the total number of requests executed.
	Requests uint64 `json:"requests"`

	// StatusCodes is a histogram of the responses' status codes.
	StatusCodes map[string]uint64 `json:"status_codes"`

	// Errors is a set of unique errors returned by the targets during the attack.
	Errors []string `json:"errors"`

	// Used for fast lookup of errors in Errors
	errors  map[string]struct{}

	// Finish message has been received
	Finished bool 
}

func (m AggMetrics) getJobID() m.JobID {
	return m.ID
}

// ProtoToIncoming Convert protobufs to Incoming type message
func ProtoToIncoming(msg *aggregator.Message) (Incoming, error) {
	var inc Incoming

	switch msg.Payload.(type) {
	case *aggregator.Message_Metrics:
		metrics := msg.GetMetrics()
		timestamp, err := pytypes.Timestamp(metrics.GetTimestamp())
		if err != nil {
			log.Fatal(err)
		}
		inc = AggMetrics{
			TestId:		metrics.GetTestId(),
			InstanceId:	metrics.GetInstanceId(),
			JobId:      m.JobID(metrics.GetJobId()),
			Requests:	metric.GetRequests(),
			Code:		metrics.GetCode(),
			BytesIn: 	metrics.GetBytesIn(),
			BytesOut:	metrics.GetBytesOut(),
			Latencies:	time.Duration(metrics.GetLatencies()),
			Earliest:	metrics.GetEarliest(),
			Latest:		metrics.GetLatest(),
			End:		metrics.GetEnd(),
			Errors:     metrics.GetErrors(),
			Finish:		metrics.GetFinish()
		}
	default:
		return nil, errors.New("Invalid Incoming aggregator message type")
	}

	return inc, nil
}