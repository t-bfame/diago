package aggregator

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	m "github.com/t-bfame/diago/pkg/model"
)

type Aggregator struct {
	
}

func Register() (incomingMsgs chan Incoming, err error) {
	incomingMsgs = make(chan Incoming)

	// consume received metrics
	go func() {
		for msg := range incomingMsgs {
			
		}
	}

	return incomingMsgs
}