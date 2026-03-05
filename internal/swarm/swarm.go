package swarm

import "sync"

type Task struct {
	URL string
}

type Result struct {
	URL    string
	Status int
}

type Swarm struct {
	Workers int
	Tasks   chan Task
	Results chan Result

	wg sync.WaitGroup
}

func New(workers int) *Swarm {

	return &Swarm{

		Workers: workers,
		Tasks:   make(chan Task, 100),
		Results: make(chan Result, 100),
	}
}

func (s *Swarm) Run(worker func(Task) Result) {

	for i := 0; i < s.Workers; i++ {

		s.wg.Add(1)

		go func() {

			defer s.wg.Done()

			for t := range s.Tasks {

				s.Results <- worker(t)

			}

		}()
	}
}

func (s *Swarm) Wait() {

	close(s.Tasks)

	s.wg.Wait()

	close(s.Results)
}
