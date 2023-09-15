package sabercache

import (
	"sabercache/cachememory"
)

type cache struct {
	cachememory   cachememory.CacheMemory
	capacity      int64
	cacheStrategy string
}

func newCache(capacity int64, cacheStrategy string) *cache {
	return &cache{
		capacity:      capacity,
		cacheStrategy: cacheStrategy,
	}
}

func (c *cache) SetWithoutTTL(key string, value ByteView) {
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

	c.cachememory.SetWithoutTTL(key, value)
}

func (c *cache) SetWithTTL(key string, value ByteView, expireSecond int64) {
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
	c.cachememory.SetWithTTL(key, value, expireSecond)
}
func (c *cache) Get(key string) (ByteView, bool) {
	if c.cachememory == nil {
		return ByteView{}, false
	}
	if v, ok := c.cachememory.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}

func (c *cache) TTL(key string) int {
	return c.TTL(key)
}
