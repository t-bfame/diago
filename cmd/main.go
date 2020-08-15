package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/scheduler"
)

func main() {
	ti := manager.Job{
		ID:       "1",
		Name:     "alpha",
		Group:    "hello-world",
		Priority: 0,
	}

	s := scheduler.NewScheduler()

	go func() {
		ch, err := s.Submit(ti)

		if err != nil {
			panic(err)
		}

		for msg := range ch {
			fmt.Println(msg)
		}

		fmt.Println("done boi")
	}()

	time.Sleep(1 * time.Second)

	go func() {
		time.Sleep(5 * time.Second)
		s.Stop(ti)
	}()

	ch2, err2 := s.Register("hello-world", scheduler.InstanceID("1"))

	if err2 != nil {
		panic(err2)
	}

	i := 0
	for {
		i++
		ch2 <- scheduler.Message{"1", "This is a message " + strconv.Itoa(i)}
		time.Sleep(500 * time.Millisecond)
	}

	close(ch2)

	// i := 0
	time.Sleep(3 * time.Second)
}
