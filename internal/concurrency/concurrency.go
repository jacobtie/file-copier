package concurrency

type Semaphore struct {
	guard chan struct{}
}

func NewSemaphore(numWorkers int) *Semaphore {
	return &Semaphore{
		guard: make(chan struct{}, numWorkers),
	}
}

func (s *Semaphore) Lock() {
	s.guard <- struct{}{}
}

func (s *Semaphore) Unlock() {
	<-s.guard
}
