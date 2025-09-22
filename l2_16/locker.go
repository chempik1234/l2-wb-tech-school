package main

// IWorkerLocker defines a mutex-like locker that helps to parallel some work
type IWorkerLocker interface {
	// Lock is the lock function - called before unlock
	Lock()
	// Unlock is the unlock function - called in defer
	Unlock()
}

// PoolWorkerLocker is a locker that allows to at most maxWorkers locked users at a time
type PoolWorkerLocker struct {
	maxWorkers uint32

	// buferrized channel with capacity = maxWorkers
	currentWorkers chan struct{}
}

// NewPoolWorkerLocker creates a new PoolWorkerLocker
func NewPoolWorkerLocker(maxWorkers uint32) *PoolWorkerLocker {
	return &PoolWorkerLocker{
		maxWorkers:     maxWorkers,
		currentWorkers: make(chan struct{}, maxWorkers),
	}
}

// Lock is the lock function - called before unlock, waits until there is room in main channel
func (p *PoolWorkerLocker) Lock() {
	p.currentWorkers <- struct{}{}
}

// Unlock is the unlock function - called in defer, frees some room in main channel
func (p *PoolWorkerLocker) Unlock() {
	<-p.currentWorkers
}
