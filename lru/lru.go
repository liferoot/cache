package lru

import (
	"container/list"

	"github.com/liferoot/cache"
)

type LRU struct {
	capacity  int
	entities  map[interface{}]*list.Element
	evictList *list.List
	evictCb   cache.EvictCallback
	getCb     cache.GetCallback
	popCb     cache.PopCallback
}

type entity struct{ key, value interface{} }

func (c *LRU) Cap() int { return c.capacity }
func (c *LRU) Len() int { return c.evictList.Len() }

func (c *LRU) Clear() {
	for k, v := range c.entities {
		delete(c.entities, k)
		if c.evictCb != nil {
			c.evictCb(k, v.Value.(*entity).value)
		}
	}
	c.evictList.Init()
}

func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.entities[key]
	return
}

func (c *LRU) Each(n int, f cache.EachCallback) {
	if c.Len() == 0 || f == nil {
		return
	}
	if c.Len() < n || n < 1 {
		n = c.Len()
	}
	var ent *entity

	for elem := c.evictList.Back(); n > 0; elem = elem.Prev() {
		ent = elem.Value.(*entity)
		f(ent.key, ent.value)
		n--
	}
}

func (c *LRU) Get(key interface{}) (value interface{}, err error) {
	if elem, ok := c.entities[key]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*entity).value, nil
	} else {
		if c.getCb != nil {
			if value, ok = c.getCb(key); ok {
				return
			}
		}
		if c.popCb != nil {
			if value, err = c.popCb(key); err == nil {
				c.insert(key, value)
			}
			return
		}
	}
	return nil, cache.ErrKeyNotFound
}

func (c *LRU) Peek(key interface{}) (interface{}, bool) {
	if elem, ok := c.entities[key]; ok {
		return elem.Value.(*entity).value, ok
	}
	return nil, false
}

func (c LRU) Put(key, value interface{}) {
	if elem, ok := c.entities[key]; ok {
		c.evictList.MoveToFront(elem)
		elem.Value.(*entity).value = value
	} else {
		c.insert(key, value)
	}
}

func (c *LRU) Remove(key interface{}) {
	if elem, ok := c.entities[key]; ok {
		c.remove(elem)
	}
}

func (c *LRU) insert(key, value interface{}) {
	if c.evictList.Len() == c.capacity {
		c.remove(c.evictList.Back())
	}
	c.entities[key] = c.evictList.PushFront(&entity{key, value})
}

func (c *LRU) remove(elem *list.Element) {
	ent := elem.Value.(*entity)
	c.evictList.Remove(elem)
	delete(c.entities, ent.key)
	if c.evictCb != nil {
		c.evictCb(ent.key, ent.value)
	}
}

func New(cap int, evict cache.EvictCallback, get cache.GetCallback, pop cache.PopCallback) cache.Cache {
	if cap < 1 {
		panic(`LRU: the cache capacity must be greater than zero.`)
	}
	return &LRU{
		capacity:  cap,
		entities:  make(map[interface{}]*list.Element),
		evictList: list.New(),
		evictCb:   evict,
		getCb:     get,
		popCb:     pop,
	}
}
