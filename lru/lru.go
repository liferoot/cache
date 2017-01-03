package lru

import (
	"github.com/liferoot/cache"
	"github.com/liferoot/linked"
)

type LRU struct {
	capacity int
	entries  map[interface{}]*linked.Node
	list     *linked.List
	evictCb  cache.EvictCallback
	omitCb   cache.OmitCallback
}

type entry struct{ key, value interface{} }

func (c *LRU) Cap() int { return c.capacity }
func (c *LRU) Len() int { return c.list.Len() }

func (c *LRU) Clear() {
	if c.evictCb != nil {
		for k, v := range c.entries {
			c.evictCb(k, v.Value.(*entry).value)
		}
	}
	c.list.Init()
	c.entries = make(map[interface{}]*linked.Node)
}

func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.entries[key]
	return
}

func (c *LRU) Each(n int, each cache.EachCallback) {
	if c.Len() == 0 || each == nil {
		return
	}
	if c.Len() < n || n < 1 {
		n = c.Len()
	}
	var e *entry

	for p := c.list.Last(); n > 0; p = p.Prev() {
		e = p.Value.(*entry)
		each(e.key, e.value)
		n--
	}
}

func (c *LRU) Get(key interface{}) (value interface{}, err error) {
	if e, ok := c.entries[key]; ok {
		c.list.Push(e)
		return e.Value.(*entry).value, nil
	} else if c.omitCb != nil {
		if value, ok = c.omitCb(key); ok {
			return
		}
	}
	return nil, cache.ErrKeyNotFound
}

func (c *LRU) Peek(key interface{}) (interface{}, bool) {
	if e, ok := c.entries[key]; ok {
		return e.Value.(*entry).value, ok
	}
	return nil, false
}

func (c *LRU) Put(key, value interface{}) {
	if e, ok := c.entries[key]; ok {
		c.list.Push(e)
		e.Value.(*entry).value = value
	} else {
		if c.list.Len() == c.capacity {
			c.remove(c.list.Last())
		}
		c.entries[key] = c.list.Push(&entry{key, value})
	}
}

func (c *LRU) Remove(key interface{}) {
	if e, ok := c.entries[key]; ok {
		c.remove(e)
	}
}

func (c *LRU) remove(node *linked.Node) {
	e := node.Detach().Value.(*entry)
	delete(c.entries, e.key)
	if c.evictCb != nil {
		c.evictCb(e.key, e.value)
	}
}

func New(cap int, evict cache.EvictCallback, omit cache.OmitCallback) *LRU {
	if cap < 1 {
		panic(`LRU: the cache capacity must be greater than zero.`)
	}
	return &LRU{
		capacity: cap,
		entries:  make(map[interface{}]*linked.Node),
		list:     new(linked.List),
		evictCb:  evict,
		omitCb:   omit,
	}
}
