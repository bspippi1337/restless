package bench

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/bspippi1337/restless/internal/core/types"
)

type Result struct {
	TotalRequests int64
	Errors        int64
	DurationMs    int64
	P50Ms         int64
	P95Ms         int64
	P99Ms         int64
}

type Config struct {
	Concurrency int
	Duration    time.Duration
	Request     types.Request
}

func Run(ctx context.Context, r engine.Runner, cfg Config) (Result, error) {
	if r == nil {
		return Result{}, errors.New("nil runner")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 5 * time.Second
	}

	deadline := time.Now().Add(cfg.Duration)

	var total int64
	var errs int64
	var mu sync.Mutex
	var durs []int64

	wg := sync.WaitGroup{}
	wg.Add(cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			defer wg.Done()
			for time.Now().Before(deadline) {
				start := time.Now()
				_, err := r.Run(ctx, cfg.Request)
				ms := time.Since(start).Milliseconds()
				atomic.AddInt64(&total, 1)
				if err != nil {
					atomic.AddInt64(&errs, 1)
					continue
				}
				mu.Lock()
				durs = append(durs, ms)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	out := Result{
		TotalRequests: total,
		Errors:        errs,
		DurationMs:    cfg.Duration.Milliseconds(),
	}
	out.P50Ms, out.P95Ms, out.P99Ms = percentiles(durs)
	return out, nil
}

func percentiles(ms []int64) (p50, p95, p99 int64) {
	if len(ms) == 0 {
		return 0, 0, 0
	}
	// simple in-place sort (small enough), avoid extra deps
	for i := 0; i < len(ms); i++ {
		for j := i + 1; j < len(ms); j++ {
			if ms[j] < ms[i] {
				ms[i], ms[j] = ms[j], ms[i]
			}
		}
	}
	get := func(q float64) int64 {
		if len(ms) == 0 {
			return 0
		}
		idx := int(float64(len(ms)-1) * q)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(ms) {
			idx = len(ms) - 1
		}
		return ms[idx]
	}
	return get(0.50), get(0.95), get(0.99)
}
