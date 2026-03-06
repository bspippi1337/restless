package status

import (
	"fmt"
	"sync"
	"time"
)

var mu sync.Mutex

var reqCount int
var epCount int
var probeCount int
var consCount int
var errCount int

var start = time.Now()

func IncRequest() {
	mu.Lock()
	reqCount++
	mu.Unlock()
}

func IncEndpoint() {
	mu.Lock()
	epCount++
	mu.Unlock()
}

func IncProbe() {
	mu.Lock()
	probeCount++
	mu.Unlock()
}

func IncConsensus() {
	mu.Lock()
	consCount++
	mu.Unlock()
}

func IncError() {
	mu.Lock()
	errCount++
	mu.Unlock()
}

func Print() {

	mu.Lock()
	req := reqCount
	ep := epCount
	probe := probeCount
	cons := consCount
	err := errCount
	mu.Unlock()

	elapsed := time.Since(start).Seconds()

	rate := float64(req) / elapsed

	fmt.Printf(
		"\rrestless | req:%d ep:%d probe:%d cons:%d err:%d rate:%.1f/s",
		req, ep, probe, cons, err, rate,
	)
}
