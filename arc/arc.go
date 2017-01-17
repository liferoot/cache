package arc

import (
	"github.com/liferoot/cache"
	"github.com/liferoot/linked"
)

type ARC struct {
	capacity int
	p        int
	entries  map[interface{}]*linked.Node
	t1       *linked.List // recent cache entries
	t2       *linked.List // frequent cache entries, referenced at least twice
	b1       *linked.List // ghost entries recently evicted from t1
	b2       *linked.List // similar ghost entries, but evicted from t2
	evictCb  cache.EvictCallback
	omitCb   cache.OmitCallback
	shadeCb  cache.ShadeCallback
}

type entry struct{ key, value interface{} }

func (c *ARC) Cap() int { return c.capacity }
func (c *ARC) Len() int { return c.t1.Len() + c.t2.Len() }

func (c *ARC) Clear() {
	if c.evictCb != nil {
		for k, v := range c.entries {
			if v.List() == c.t1 || v.List() == c.t2 {
				c.evictCb(k, v.Value.(*entry).value)
			}
		}
	}
	c.p = 0
	c.entries = make(map[interface{}]*linked.Node)
	c.t1.Init()
	c.t2.Init()
	c.b1.Init()
	c.b2.Init()
}

func (c *ARC) Contains(key interface{}) bool {
	if e, ok := c.entries[key]; ok {
		return e.List() == c.t1 || e.List() == c.t2
	}
	return false
}

func (c *ARC) Each(n int, each cache.EachCallback) {
	z := c.Len()

	if z == 0 || each == nil {
		return
	}
	if z < n || n < 1 {
		n = z
	}
	var e *entry

	for p := c.t1.Last(); n > 0 && p != nil; p = p.Prev() {
		e = p.Value.(*entry)
		each(e.key, e.value)
		n--
	}
	for p := c.t2.Last(); n > 0 && p != nil; p = p.Prev() {
		e = p.Value.(*entry)
		each(e.key, e.value)
		n--
	}
}

func (c *ARC) Get(key interface{}) (value interface{}, err error) {
	if e, ok := c.entries[key]; ok {
		if e.List() == c.t1 || e.List() == c.t2 {
			c.t2.Push(e)
			return e.Value.(*entry).value, nil
		}
	} else if c.omitCb != nil {
		if value, ok = c.omitCb(key); ok {
			return
		}
	}
	return nil, cache.ErrKeyNotFound
}

func (c *ARC) Peek(key interface{}) (interface{}, bool) {
	if e, ok := c.entries[key]; ok && (e.List() == c.t1 || e.List() == c.t2) {
		return e.Value.(*entry).value, true
	}
	return nil, false
}

func (c *ARC) Put(key, value interface{}) {
	if e, ok := c.entries[key]; ok {
		switch e.List() {
		case c.b1:
			c.p = min(c.capacity, c.p+max(1, c.b2.Len()/c.b1.Len()))
			c.replace(c.b1)
		case c.b2:
			c.p = max(0, c.p-max(1, c.b1.Len()/c.b2.Len()))
			c.replace(c.b2)
		}
		c.t2.Push(e)
		e.Value.(*entry).value = value
	} else {
		if z := c.t1.Len() + c.b1.Len(); z == c.capacity {
			if c.t1.Len() < c.capacity {
				c.remove(c.b1.Last())
				c.replace(c.b1)
			} else {
				c.remove(c.t1.Last())
			}
		} else if zz := z + c.t2.Len() + c.b2.Len(); z < c.capacity && zz >= c.capacity {
			if zz == c.capacity<<1 {
				c.remove(c.b2.Last())
			}
			c.replace(nil)
		}
		c.entries[key] = c.t1.Push(&entry{key, value})
	}
}

func (c *ARC) Remove(key interface{}) {
	if e, ok := c.entries[key]; ok {
		c.remove(e)
	}
}

func (c *ARC) remove(node *linked.Node) {
	e := node.Value.(*entry)
	delete(c.entries, e.key)
	if c.evictCb != nil && (node.List() == c.t1 || node.List() == c.t2) {
		c.evictCb(e.key, e.value)
	}
	node.Detach()
}

func (c *ARC) replace(l *linked.List) {
	var node *linked.Node

	if c.t1.Len() > 0 && (c.t1.Len() > c.p || (l == c.b2 && c.t1.Len() == c.p)) {
		node = c.t1.Last()
		c.b1.Push(node)
	} else {
		node = c.t2.Last()
		c.b2.Push(node)
	}
	e := node.Value.(*entry)

	if c.shadeCb != nil {
		c.shadeCb(e.key, e.value)
	}
	e.value = nil
}

func New(cap int, evict cache.EvictCallback, omit cache.OmitCallback, shade cache.ShadeCallback) *ARC {
	if cap < 1 {
		panic(`ARC: the cache capacity must be greater than zero.`)
	}
	return &ARC{
		capacity: cap,
		entries:  make(map[interface{}]*linked.Node),
		t1:       new(linked.List),
		t2:       new(linked.List),
		b1:       new(linked.List),
		b2:       new(linked.List),
		evictCb:  evict,
		omitCb:   omit,
		shadeCb:  shade,
	}
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
