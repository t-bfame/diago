package main

import (
	"fmt"
	"time"

	"github.com/t-bfame/diago/internal/scheduler"
)

func main() {
	fmt.Println("hello world 3")

	s := scheduler.NewScheduler()

	for {
		s.Schedule()

		time.Sleep(10 * time.Second)
	}
}
