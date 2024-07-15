package utils

import (
	"sync"
	"time"
)

type IClock interface {
	Now() time.Time
}

type Clock struct{}

func (c *Clock) Now() time.Time {
	return time.Now()
}

type MockClock struct {
	mutex sync.RWMutex
	now   time.Time
}

func (c *MockClock) SetTime(tm time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = tm
}

func (c *MockClock) Now() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.now
}

type BrokenClock struct{}

func (c *BrokenClock) Now() time.Time {
	panic("This timer is broken by design (for tests only)")
}
