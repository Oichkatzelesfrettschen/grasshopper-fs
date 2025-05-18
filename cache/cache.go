package cache

import (
	"container/list"
	"sync"

	"github.com/mit-pdos/go-journal/util"
)

type Cslot[T any] struct {
	Obj T
}

type entry[T any] struct {
	slot Cslot[T]
	lru  *list.Element
	id   uint64
}

type Cache[T any] struct {
	mu      *sync.Mutex
	entries map[uint64]*entry[T]
	lru     *list.List
	sz      uint64
	cnt     uint64
}

func MkCache[T any](sz uint64) *Cache[T] {
	entries := make(map[uint64]*entry[T], sz)
	return &Cache[T]{
		mu:      new(sync.Mutex),
		entries: entries,
		lru:     list.New(),
		cnt:     0,
		sz:      sz,
	}
}

func (c *Cache[T]) PrintCache() {
	for k, v := range c.entries {
		util.DPrintf(0, "Entry %v %v\n", k, v)
	}
}

func (c *Cache[T]) evict() {
	e := c.lru.Front()
	if e == nil {
		c.PrintCache()
		panic("evict")
	}
	entry := e.Value.(*entry[T])
	c.lru.Remove(e)
	util.DPrintf(5, "evict: %d\n", entry.id)
	delete(c.entries, entry.id)
	c.cnt = c.cnt - 1
}

func (c *Cache[T]) LookupSlot(id uint64) *Cslot[T] {
	c.mu.Lock()
	e := c.entries[id]
	if e != nil {
		if id != e.id {
			panic("LookupSlot")
		}
		if e.lru != nil {
			c.lru.Remove(e.lru)
			e.lru = c.lru.PushBack(e)
		}
		c.mu.Unlock()
		return &e.slot
	}
	if c.cnt >= c.sz {
		c.evict()
	}
	var zero T
	enew := &entry[T]{
		slot: Cslot[T]{Obj: zero},
		lru:  nil,
		id:   id,
	}
	c.entries[id] = enew
	enew.lru = c.lru.PushBack(enew)
	c.cnt = c.cnt + 1
	c.mu.Unlock()
	return &enew.slot
}
