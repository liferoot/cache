package cache

import "errors"

type Cache interface {
	Cap() int
	Len() int
	Clear()
	Contains(key interface{}) bool
	Each(n int, each EachCallback)
	Get(key interface{}) (value interface{}, err error)
	Peek(key interface{}) (value interface{}, ok bool)
	Put(key, value interface{})
	Remove(key interface{})
}

type EachCallback func(key, value interface{})
type EvictCallback func(key, value interface{})
type OmitCallback func(key interface{}) (value interface{}, ok bool)
type ShadeCallback func(key, value interface{})

var (
	ErrKeyNotFound = errors.New(`key not found`)
)
