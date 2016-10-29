package lru

import (
	"reflect"
	"testing"

	"github.com/liferoot/cache"
)

func TestLen(t *testing.T) {
	c := New(2, nil, nil)

	if c.Put(1, 1); c.Len() != 1 {
		t.Errorf(`expected cache length 1, got %d\n`, c.Len())
	}
	if c.Put(2, 2); c.Len() != 2 {
		t.Errorf(`expected cache length 2, got %d\n`, c.Len())
	}
	if c.Put(3, 3); c.Len() != 2 {
		t.Errorf(`expected сache length 2, got %d\n`, c.Len())
	}
}

func TestContains(t *testing.T) {
	c := New(1, func(k, v interface{}) {
		if k != 1 {
			t.Errorf("evict: expect key 1, got %v\n", k)
		}
	}, nil)

	if c.Put(1, 1); !c.Contains(1) {
		t.Error(`key 1 does not exist`)
	}
	if c.Put(2, 2); !c.Contains(2) {
		t.Error(`key 2 does not exist`)
	}
	if c.Contains(1) {
		t.Error(`key 1 exists`)
	}
}

func TestGet(t *testing.T) {
	evict := 1
	c := New(3,
		func(k, v interface{}) {
			if k != evict {
				t.Errorf("evict: expected key %v, got %v\n", evict, k)
			}
			evict += 2
		},
		func(k interface{}) (interface{}, error) {
			if k == 4 {
				return 4, nil
			}
			return nil, cache.ErrKeyNotFound
		})
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)

	for i := 1; i < 5; i++ {
		if v, e := c.Get(i); v != i || e != nil {
			t.Errorf("\n\tfor: %v\n\texp: %v, err: <nil>\n\tgot: %v, err: %v\n", i, i, v, e)
		}
	}
	if v, e := c.Get(1); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("\n\tfor: 1\n\texp: 1, err: %v\n\tgot: %v, err: %v\n", cache.ErrKeyNotFound, v, e)
	}
	if v, e := c.Get(2); v != 2 || e != nil {
		t.Errorf("\n\tfor: 2\n\texp: 2, err: <nil>\n\tgot: %v, err: %v\n", v, e)
	}
	if v, e := c.Get(5); v != nil || e != cache.ErrKeyNotFound {
		t.Errorf("\n\tfor: 5\n\texp: 5, err: %v\n\tgot: %v, err: %v\n", cache.ErrKeyNotFound, v, e)
	}
	c.Put(5, 5)

	if v, e := c.Get(5); v != 5 || e != nil {
		t.Errorf("\n\tfor: 5\n\texp: 5, err: <nil>\n\tgot: %v, err: %v\n", v, e)
	}
}

func TestPeek(t *testing.T) {
	c := New(1, nil, nil)
	c.Put(1, 1)

	if v, ok := c.Peek(1); !ok || v != 1 {
		t.Errorf("\n\tfor: 1\n\texp: 1, ok: true\n\tgot: %v, ok: %t\n", v, ok)
	}
	if v, ok := c.Peek(2); ok || v != nil {
		t.Errorf("\n\tfor: 2\n\texp: <nil>, ok: false\n\tgot: %v, ok: %t\n", v, ok)
	}
}

func TestRemove(t *testing.T) {
	c := New(1, nil, nil)
	c.Put(1, 1)
	c.Remove(1)

	if c.Contains(1) {
		t.Error(`key 1 exists`)
	}
}

func TestEach(t *testing.T) {
	n := 8
	c := New(n, nil, nil)
	exp, out := make([]int, n), make([]int, 0, n)

	for i := 0; i < n; i++ {
		c.Put(i, i)
		exp[i] = i
	}
	c.Each(0, func(k, v interface{}) {
		out = append(out, k.(int))
	})
	if !reflect.DeepEqual(exp, out) {
		t.Errorf("\n\texp: %v\n\tgot: %v\n", exp, out)
	}
}
