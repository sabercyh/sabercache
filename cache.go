package sabercache

import (
	"sabercache/cachememory"
	"sync"
)

type cache struct {
	mu            sync.RWMutex
	cachememory   cachememory.CacheMemory
	capacity      int64
	cacheStrategy string
}

func newCache(capacity int64, cacheStrategy string) *cache {
	c := &cache{
		capacity:      capacity,
		cacheStrategy: cacheStrategy,
	}

	return c
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachememory == nil {
		switch {
		case c.cacheStrategy == "lfu":
			c.cachememory = cachememory.NewLFUCache(c.capacity, nil)
		case c.cacheStrategy == "fifo":
			c.cachememory = cachememory.NewFIFOCache(c.capacity, nil)
		case c.cacheStrategy == "lru":
			c.cachememory = cachememory.NewLRUCache(c.capacity, nil)
		default:
			c.cachememory = cachememory.NewLFUCache(c.capacity, nil)
		}
	}

	c.cachememory.Add(key, value)
}
func (c *cache) get(key string) (ByteView, bool) {
	if c.cachememory == nil {
		return ByteView{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cachememory.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}
