package main

import (
	"fmt"

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
	s.Schedule(ti)

}
