package aggregator

<<<<<<< Updated upstream
import "fmt"

// "errors"
// "sync"

// log "github.com/sirupsen/logrus"
// m "github.com/t-bfame/diago/pkg/model"

type Aggregator struct {
=======
type Aggregator struct {
	outputChannels map[TestId]map[InstanceId]map[JobId]chan
}

func RegisterJob(testId TestId, instanceId InstanceId, jobId JobId) {
	output = make(chan AggMetrics)
	// TODO: do we need to create empty maps or something?
	outputChannels[testId][instanceId][jobId] = output
	return output
>>>>>>> Stashed changes
}

func Register() (incomingMsgs chan Incoming, err error) {
	incomingMsgs = make(chan Incoming)

	// consume received metrics
	go func() {
		for msg := range incomingMsgs {
<<<<<<< Updated upstream
			fmt.Println(msg)
		}
	}()

	return incomingMsgs, nil
=======
			if msg.Finished {
				// trigger podgrp worker deletion (podgrp.go#L200)
			}
			outputChannels[msg.TestId][msg.InstanceId][msg.JobId] <- msg
		}
	}()

	return incomingMsgs
>>>>>>> Stashed changes
}
