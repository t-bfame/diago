package main

import (
	"fmt"
	"time"

	"github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/scheduler"
)

func main() {
	fmt.Println("hello world 3")

	envs := map[string]string{
		"rsc": "a",
	}

	ti := manager.TestInstance{
		Id:       "1",
		Name:     "alpha",
		Image:    "hello-world",
		Priority: 0,
		Env:      envs,
	}

	s := scheduler.NewScheduler()
	id, err := s.Schedule(ti)

	if err != nil {
		panic(err)
	}

	i := 0

	for {
		time.Sleep(10 * time.Second)
		i++

		if i == 3 {
			s.Unschedule(ti, id)
		}
	}
}
