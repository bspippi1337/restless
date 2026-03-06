package ui

import (
	"fmt"
	"sync/atomic"
	"time"
)

var req uint64
var endpoints uint64

func IncRequest() {
	atomic.AddUint64(&req, 1)
}

func IncEndpoint() {
	atomic.AddUint64(&endpoints, 1)
}

func StartStatus() {

	go func() {

		start := time.Now()

		for {

			time.Sleep(1 * time.Second)

			r := atomic.LoadUint64(&req)
			e := atomic.LoadUint64(&endpoints)

			elapsed := time.Since(start).Seconds()

			rate := float64(r) / elapsed

			fmt.Printf("\rrestless | req:%d ep:%d rate:%.1f/s", r, e, rate)

		}

	}()

}
