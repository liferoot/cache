package lru

import (
	"reflect"
	"testing"

	"github.com/liferoot/cache"
	"github.com/liferoot/linked"
)

func TestNewWithPanic(t *testing.T) {
	defer func() {
		if e := recover(); e == nil {
			t.Error("panic did not occur")
		}
	}()
	New(0, nil, nil)
}

func TestCap(t *testing.T) {
	c := New(1024, nil, nil)
	if c.Cap() != 1024 {
		t.Errorf("expected capacity 1024, got %d", c.Cap())
	}
}

func TestLen(t *testing.T) {
	c := New(2, nil, nil)
	chklen(t, `case 1`, c, 0)
	c.Put(1, 1)
	chklen(t, `case 2`, c, 1)
	c.Put(2, 2)
	chklen(t, `case 3`, c, 2)
	c.Put(3, 3)
	chklen(t, `case 4`, c, 2)
}

func TestClear(t *testing.T) {
	c := New(3, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)
	chklen(t, `case 1`, c, 3)
	c.Clear()
	chklen(t, `case 2`, c, 0)
}

func TestClearWithEvict(t *testing.T) {
	exp := []interface{}{1, 2, 3}
	l := len(exp)
	c := New(3, func(k, v interface{}) {
		ok := false
		for _, val := range exp {
			if val == v {
				ok = true
				l--
			}
		}
		if !ok {
			t.Errorf("unexpected value %v", v)
		}
	}, nil)
	for _, v := range exp {
		c.Put(v, v)
	}
	chklen(t, `case 1`, c, l)
	if c.Clear(); l > 0 {
		t.Errorf("expected number of eviction %d, got %d", len(exp), l)
	}
	chklen(t, `case 2`, c, 0)
}

func TestContains(t *testing.T) {
	c := New(2, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	if c.Contains(3) {
		t.Error(`key 3 exists`)
	}
	if !c.Contains(2) {
		t.Error(`key 2 does not exist`)
	}
	if !c.Contains(1) {
		t.Error(`key 1 does not exist`)
	}
	chkhead(t, ``, c.list, 2, 2)
}

func TestGet(t *testing.T) {
	c := New(2, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	if v, e := c.Get(1); v != 1 || e != nil {
		t.Errorf("expected (1, <nil>), got (%v, %v)", v, e)
	}
	chkhead(t, `case 1`, c.list, 1, 1)
	if v, e := c.Get(2); v != 2 || e != nil {
		t.Errorf("expected (2, <nil>), got (%v, %v)", v, e)
	}
	if v, e := c.Get(3); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("expected (<nil>, %v), got (%v, %v)", cache.ErrKeyNotFound, v, e)
	}
	chkhead(t, `case 2`, c.list, 2, 2)
}

func TestGetWithOmit(t *testing.T) {
	c := New(1, nil, func(k interface{}) (interface{}, bool) {
		if k == 2 {
			return 2, true
		}
		return nil, false
	})
	if v, e := c.Get(1); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("expected (<nil>, %v), got (%v, %v)", cache.ErrKeyNotFound, v, e)
	}
	if v, e := c.Get(2); v != 2 || e != nil {
		t.Errorf("expected (2, <nil>), got (%v, %v)", v, e)
	}
}

func TestPeek(t *testing.T) {
	c := New(2, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	if v, ok := c.Peek(1); !ok || v != 1 {
		t.Errorf("expected (1, true), got (%v, %t)", v, ok)
	}
	if v, ok := c.Peek(3); ok || v != nil {
		t.Errorf("expected (nil, false), got (%v, %t)", v, ok)
	}
	chkhead(t, ``, c.list, 2, 2)
}

func TestPut(t *testing.T) {
	c := New(2, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	chkhead(t, `case 1`, c.list, 2, 2)
	c.Put(1, 11)
	chkhead(t, `case 2`, c.list, 1, 11)
}

func TestRemove(t *testing.T) {
	c := New(1, nil, nil)
	c.Put(1, 1)
	if c.Remove(1); c.Contains(1) {
		t.Error(`key 1 exists`)
	}
	chklen(t, ``, c, 0)
}

func TestRemoveWithEvict(t *testing.T) {
	c := New(1, func(k, v interface{}) {
		if v != 1 {
			t.Errorf("expected value for eviction 1, got %v", v)
		}
	}, nil)
	c.Put(1, 1)
	c.Remove(1)
	chklen(t, ``, c, 0)
}

func TestEach(t *testing.T) {
	n := 8
	c := New(n, nil, nil)
	exp, out := make([]interface{}, n), make([]interface{}, 0, n)

	c.Each(0, func(k, v interface{}) {
		t.Error(`aasd`)
	})
	for i := 0; i < n; i++ {
		c.Put(i, i)
		exp[i] = i
	}
	c.Each(0, func(k, v interface{}) {
		out = append(out, k)
	})
	if !reflect.DeepEqual(exp, out) {
		t.Errorf("\n\texp: %v\n\tgot: %v\n", exp, out)
	}
}

func chklen(t *testing.T, prefix string, c *LRU, length int) (ok bool) {
	if len(prefix) > 0 {
		prefix += `: `
	}
	if ok = c.list.Len() == len(c.entries); !ok {
		t.Errorf("%slength of the internal list and map are not equal", prefix)
		return
	}
	if ok = c.Len() == length; !ok {
		t.Errorf("%sexpected length %d, got %d", prefix, length, c.Len())
	}
	return
}

func chkhead(t *testing.T, prefix string, l *linked.List, key, value interface{}) {
	if len(prefix) > 0 {
		prefix += `: `
	}
	if first := l.First(); first == nil {
		t.Error(`internal list is empty`)
	} else {
		e := first.Value.(*entry)
		if e.key != key {
			t.Errorf("expected key %v, got %v", key, e.key)
		}
		if e.value != value {
			t.Errorf("expected value %v, got %v", value, e.value)
		}
	}
}
