package telemetry

import (
	"fmt"
	"sync"
	"time"
)

type Engine struct {
	mu sync.Mutex

	start time.Time

	Requests  int
	Endpoints int
	Probes    int
	Consensus int
	Queue     int
	Workers   int
	Errors    int
}

var T = &Engine{
	start: time.Now(),
}

func IncRequest() {
	T.mu.Lock()
	T.Requests++
	T.mu.Unlock()
}

func IncEndpoint() {
	T.mu.Lock()
	T.Endpoints++
	T.mu.Unlock()
}

func IncProbe() {
	T.mu.Lock()
	T.Probes++
	T.mu.Unlock()
}

func IncConsensus() {
	T.mu.Lock()
	T.Consensus++
	T.mu.Unlock()
}

func IncError() {
	T.mu.Lock()
	T.Errors++
	T.mu.Unlock()
}

func SetQueue(n int) {
	T.mu.Lock()
	T.Queue = n
	T.mu.Unlock()
}

func SetWorkers(n int) {
	T.mu.Lock()
	T.Workers = n
	T.mu.Unlock()
}

func Print() {

	T.mu.Lock()
	defer T.mu.Unlock()

	rate := float64(T.Requests) / time.Since(T.start).Seconds()

	fmt.Printf(
		"\rrestless | req:%d ep:%d probe:%d cons:%d queue:%d workers:%d err:%d rate:%.1f/s",
		T.Requests,
		T.Endpoints,
		T.Probes,
		T.Consensus,
		T.Queue,
		T.Workers,
		T.Errors,
		rate,
	)
}

func Start() {

	go func() {

		t := time.NewTicker(1 * time.Second)

		for range t.C {
			Print()
		}

	}()

}
