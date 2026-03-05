package status

import (
	"fmt"
	"sync"
	"time"
)

type GlobalStatus struct {
	mu        sync.Mutex
	Start     time.Time
	Requests  int
	Endpoints int
	Probes    int
	Consensus int
}

var S = &GlobalStatus{
	Start: time.Now(),
}

func IncRequest() {
	S.mu.Lock()
	S.Requests++
	S.mu.Unlock()
}

func IncEndpoint() {
	S.mu.Lock()
	S.Endpoints++
	S.mu.Unlock()
}

func IncProbe() {
	S.mu.Lock()
	S.Probes++
	S.mu.Unlock()
}

func IncConsensus() {
	S.mu.Lock()
	S.Consensus++
	S.mu.Unlock()
}

func Print() {

	S.mu.Lock()
	defer S.mu.Unlock()

	rate := float64(S.Requests) / time.Since(S.Start).Seconds()

	fmt.Printf(
		"\rrestless  req:%d  endpoints:%d  probes:%d  consensus:%d  rate:%.1f/s",
		S.Requests,
		S.Endpoints,
		S.Probes,
		S.Consensus,
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
