package junk

import "sync"

type changeClock uint64

type ChangeTracker struct {
	lock   *sync.Mutex
	notice *sync.Cond
	epoch  changeClock
}

func NewChangeTracker() *ChangeTracker {
	out := &ChangeTracker{
		lock:  &sync.Mutex{},
		epoch: 0,
	}
	out.notice = sync.NewCond(out.lock)
	return out
}

func (c *ChangeTracker) Fork() *ChangePoint {
	c.lock.Lock()
	defer c.lock.Unlock()
	return &ChangePoint{
		tracker:    c,
		afterEpoch: c.epoch,
	}
}

func (c *ChangeTracker) Update() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.epoch++
	c.notice.Broadcast()
}

type ChangePoint struct {
	tracker    *ChangeTracker
	afterEpoch changeClock
}

func (c *ChangePoint) Wait() {
	c.tracker.lock.Lock()
	defer c.tracker.lock.Unlock()
	for c.tracker.epoch <= c.afterEpoch {
		c.tracker.notice.Wait()
	}
}
