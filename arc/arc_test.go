package arc

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
	New(0, nil, nil, nil)
}

func TestCap(t *testing.T) {
	c := New(1024, nil, nil, nil)
	if c.Cap() != 1024 {
		t.Errorf("expected capacity 1024, got %d", c.Cap())
	}
}

func TestLen(t *testing.T) {
	c := New(2, nil, nil, nil)
	chklen(t, `case 1`, c, 0)
	c.Put(1, 1)
	chklen(t, `case 2`, c, 1)
	c.Put(2, 2)
	chklen(t, `case 3`, c, 2)
	c.Put(3, 3)
	chklen(t, `case 4`, c, 2)
	c.Put(2, 22)
	chklen(t, `case 5`, c, 2)
}

func TestClear(t *testing.T) {
	c := New(3, nil, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)
	chklen(t, `case 1`, c, 3)
	c.Clear()
	chklen(t, `case 2`, c, 0)
}

func TestClearWithEvict(t *testing.T) {
	exp := []interface{}{1, 2, 3, 4}
	l := len(exp)
	c := New(l, func(k, v interface{}) {
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
	}, nil, nil)
	for i, v := range exp {
		if i >= l>>1 {
			c.Put(v, v)
		}
		c.Put(v, v)
	}
	chklen(t, `case 1`, c, l)
	if c.Clear(); l > 0 {
		t.Errorf("expected number of eviction %d, got %d", len(exp), len(exp)-l)
	}
	chklen(t, `case 2`, c, 0)
}

func TestContains(t *testing.T) {
	c := New(2, nil, nil, nil)
	c.Put(1, 1)
	c.Put(1, 11)
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
	chkhead(t, ``, c.t1, 2, 2)
}

func TestGet(t *testing.T) {
	c := New(3, nil, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)
	if v, e := c.Get(1); v != 1 || e != nil {
		t.Errorf("expected (1, <nil>), got (%v, %v)", v, e)
	}
	chkhead(t, `case 1`, c.t2, 1, 1)
	if v, e := c.Get(2); v != 2 || e != nil {
		t.Errorf("expected (2, <nil>), got (%v, %v)", v, e)
	}
	if v, e := c.Get(4); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("expected (<nil>, %v), got (%v, %v)", cache.ErrKeyNotFound, v, e)
	}
	chkhead(t, `case 2`, c.t2, 2, 2)
	chkhead(t, `case 3`, c.t1, 3, 3)
}

func TestGetWithOmit(t *testing.T) {
	c := New(1, nil, func(k interface{}) (interface{}, bool) {
		if k == 2 {
			return 2, true
		}
		return nil, false
	}, nil)
	if v, e := c.Get(1); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("expected (<nil>, %v), got (%v, %v)", cache.ErrKeyNotFound, v, e)
	}
	if v, e := c.Get(2); v != 2 || e != nil {
		t.Errorf("expected (2, <nil>), got (%v, %v)", v, e)
	}
}

func TestPeek(t *testing.T) {
	c := New(2, nil, nil, nil)
	c.Put(1, 1)
	c.Put(2, 2)
	if v, ok := c.Peek(1); !ok || v != 1 {
		t.Errorf("expected (1, true), got (%v, %t)\n", v, ok)
	}
	if v, ok := c.Peek(3); ok || v != nil {
		t.Errorf("expected (<nil>, false), got (%v, %t)\n", v, ok)
	}
	chkhead(t, ``, c.t1, 2, 2)
}

func TestPut(t *testing.T) {
	i, g := 0, [...]interface{}{2, 11, 3, 22, 4, 111}
	c := New(2, nil, nil, func(k, v interface{}) {
		if v != g[i] {
			t.Errorf("expected ghost value %v, got %v", g[i], v)
		}
		i++
	})
	c.Put(1, 1)
	c.Put(2, 2)
	chkhead(t, `case 1`, c.t1, 2, 2)
	c.Put(1, 11)
	chkhead(t, `case 2`, c.t2, 1, 11)
	c.Put(3, 3)
	chkhead(t, `case 3`, c.b1, 2, nil)
	c.Put(2, 22)
	chkhead(t, `case 4`, c.b2, 1, nil)
	c.Put(1, 111)
	chkhead(t, `case 5`, c.b1, 3, nil)
	c.Put(4, 4)
	c.Put(5, 5)
	chkhead(t, `case 6/t1`, c.t1, 5, 5)
	chkhead(t, `case 6/t2`, c.t2, 1, 111)
	chkhead(t, `case 6/b1`, c.b1, 4, nil)
	chkhead(t, `case 6/b2`, c.b2, 2, nil)
	c.Put(5, 55)
	c.Put(6, 6)
	chkhead(t, `case 7/t1`, c.t1, 6, 6)
	chkhead(t, `case 7/t2`, c.t2, 5, 55)
	chkhead(t, `case 7/b1`, c.b1, 4, nil)
	chkhead(t, `case 7/b2`, c.b2, 1, nil)
}

func TestRemove(t *testing.T) {
	c := New(2, nil, nil, nil)
	c.Put(1, 1)
	c.Put(1, 11)
	c.Put(2, 2)
	if c.Remove(1); c.Contains(1) {
		t.Error(`key 1 exists`)
	}
	if c.Remove(2); c.Contains(2) {
		t.Error(`key 2 exists`)
	}
	chklen(t, ``, c, 0)
}

func TestRemoveWithEvict(t *testing.T) {
	c := New(1, func(k, v interface{}) {
		if v != 1 {
			t.Errorf("expected value for eviction 1, got %v", v)
		}
	}, nil, nil)
	c.Put(1, 1)
	c.Remove(1)
	chklen(t, ``, c, 0)
}

func TestEach(t *testing.T) {
	n := 8
	c := New(n, nil, nil, nil)
	exp, out := make([]interface{}, n), make([]interface{}, 0, n)

	c.Each(0, func(k, v interface{}) {
		t.Errorf("unexpected cache entry (%v, %v)", k, v)
	})
	for i, n2 := 0, n>>1; i < n; i++ {
		if i >= n2 {
			c.Put(i, i)
		}
		c.Put(i, i)
		exp[i] = i
	}
	c.Each(0, func(k, v interface{}) {
		out = append(out, v)
	})
	if !reflect.DeepEqual(exp, out) {
		t.Errorf("\n\texp: %v\n\tgot: %v\n", exp, out)
	}
}

func chklen(t *testing.T, prefix string, c *ARC, length int) (ok bool) {
	if len(prefix) > 0 {
		prefix += `: `
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
		t.Errorf(`%sinternal list is empty`, prefix)
	} else {
		e := first.Value.(*entry)
		if e.key != key {
			t.Errorf("%sexpected key %v, got %v", prefix, key, e.key)
		}
		if e.value != value {
			t.Errorf("%sexpected value %v, got %v", prefix, value, e.value)
		}
	}
}
