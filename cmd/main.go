package main

import (
	"fmt"

	"github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/scheduler"
)

func main() {
	fmt.Println("hello world 3")

	ti := manager.Job{
		ID:       "1",
		Name:     "alpha",
		Group:    "hello-world",
		Priority: 0,
	}

	s := scheduler.NewScheduler()
	ch, err := s.Submit(ti)

	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range ch {
			fmt.Println(msg)
		}
	}()

	// i := 0

	// for {
	// 	time.Sleep(10 * time.Second)
	// 	i++

	// 	if i == 3 {
	// 		s.Unschedule(ti, id)
	// 	}
	// }
}
