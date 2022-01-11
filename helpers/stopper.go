package helpers

type Stopper struct {
	s chan struct{}
}

func NewStopper() Stopper {
	return Stopper{s: make(chan struct{})}
}

func (s Stopper) Stop() {
	close(s.s)
}

func (s Stopper) Wait() {
	<-s.s
}
