package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct {
	ID       int64
	Requests int64
	Last     int64 // time in ns
}

func (c *Counter) Inc() {
	atomic.AddInt64(&c.Requests, 1)
	now := time.Now().UnixNano()
	for {
		old := atomic.LoadInt64(&c.Last)
		if now <= old {
			break
		}
		if atomic.CompareAndSwapInt64(&c.Last, old, now) {
			break
		}
	}
}

type Stats struct {
	lock sync.RWMutex
	data map[int64]*Counter
}

func (s *Stats) Inc(id int64) {
	if !s.fastInc(id) {
		s.slowInc(id)
	}
}

func (s *Stats) Pop() []*Counter {
	s.lock.Lock()
	cp := s.data
	s.data = nil
	s.lock.Unlock()

	var v = make([]*Counter, 0, len(cp))
	for _, c := range cp {
		v = append(v, c)
	}
	return v
}

func (s *Stats) fastInc(id int64) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.data[id]
	if ok {
		v.Inc()
	}
	return ok
}

func (s *Stats) slowInc(id int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.data[id]
	if ok {
		v.Inc()
		return
	}
	if s.data == nil {
		s.data = map[int64]*Counter{}
	}
	v = &Counter{ID: id, Last: time.Now().UnixNano(), Requests: 1}
	s.data[id] = v
}

type Cached[T any] struct {
	value T
	ready bool
	lock  sync.RWMutex
}

func (c *Cached[T]) Get(factory func() (T, error)) (T, error) { //nolint:ireturn
	c.lock.RLock()
	if c.ready {
		c.lock.RUnlock()
		return c.value, nil
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.ready {
		return c.value, nil
	}

	d, err := factory()
	if err != nil {
		return c.value, err
	}
	c.value = d
	c.ready = true
	return d, nil
}
