package progress

import (
	"fmt"
	"sync"
	"time"
)

type DiscoveryProgress struct {
	mu        sync.Mutex
	Start     time.Time
	Requests  int
	Endpoints int
	Queue     int
}

func NewDiscoveryProgress() *DiscoveryProgress {
	return &DiscoveryProgress{
		Start: time.Now(),
	}
}

func (p *DiscoveryProgress) IncRequest() {
	p.mu.Lock()
	p.Requests++
	p.mu.Unlock()
}

func (p *DiscoveryProgress) IncEndpoint() {
	p.mu.Lock()
	p.Endpoints++
	p.mu.Unlock()
}

func (p *DiscoveryProgress) SetQueue(n int) {
	p.mu.Lock()
	p.Queue = n
	p.mu.Unlock()
}

func (p *DiscoveryProgress) Print() {
	p.mu.Lock()
	defer p.mu.Unlock()

	rate := float64(p.Requests) / time.Since(p.Start).Seconds()

	fmt.Printf(
		"\rdiscover  req:%d  endpoints:%d  queue:%d  rate:%.1f/s",
		p.Requests,
		p.Endpoints,
		p.Queue,
		rate,
	)
}

func (p *DiscoveryProgress) StartTicker() {

	go func() {

		t := time.NewTicker(1 * time.Second)

		for range t.C {
			p.Print()
		}

	}()
}
