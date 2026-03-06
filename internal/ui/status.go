package ui

import (
	"fmt"
	"sync/atomic"
	"time"
)

var req uint64
var ep uint64

func IncRequest()  { atomic.AddUint64(&req, 1) }
func IncEndpoint() { atomic.AddUint64(&ep, 1) }

func Start() {

	start := time.Now()

	go func() {

		for {

			time.Sleep(time.Second)

			r := atomic.LoadUint64(&req)
			e := atomic.LoadUint64(&ep)

			elapsed := time.Since(start).Seconds()

			rate := float64(r) / elapsed

			fmt.Printf(
				"\rrestless | req:%d ep:%d rate:%.1f/s",
				r, e, rate)

		}

	}()

}
