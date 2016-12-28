package cache

import "sync"

func Safe(c Cache) Cache { return &safe{c: c} }

type safe struct {
	mu sync.RWMutex
	c  Cache
}

func (s *safe) Cap() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Cap()
}

func (s *safe) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Len()
}

func (s *safe) Clear() {
	s.mu.Lock()
	s.c.Clear()
	s.mu.Unlock()
}

func (s *safe) Contains(key interface{}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Contains(key)
}

func (s *safe) Each(n int, each EachCallback) {
	s.mu.Lock()
	s.c.Each(n, each)
	s.mu.Unlock()
}

func (s *safe) Get(key interface{}) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.Get(key)
}

func (s *safe) Peek(key interface{}) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.c.Peek(key)
}

func (s *safe) Put(key, value interface{}) {
	s.mu.Lock()
	s.c.Put(key, value)
	s.mu.Unlock()
}

func (s *safe) Remove(key interface{}) {
	s.mu.Lock()
	s.c.Remove(key)
	s.mu.Unlock()
}
