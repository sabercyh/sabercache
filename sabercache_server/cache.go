package sabercache_server

import (
	"sabercache_server/cachememory"
)

type Cache struct {
	cachememory   cachememory.CacheMemory
	capacity      int64
	cacheStrategy string
}

func newCache(capacity int64, cacheStrategy string) *Cache {
	return &Cache{
		capacity:      capacity,
		cacheStrategy: cacheStrategy,
	}
}

func (c *Cache) SetWithoutTTL(key string, value ByteView) {
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

func (c *Cache) SetWithTTL(key string, value ByteView, ttl int64) {
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
	c.cachememory.SetWithTTL(key, value, ttl)
}
func (c *Cache) Get(key string) (ByteView, bool) {
	if c.cachememory == nil {
		return ByteView{}, false
	}
	if v, ok := c.cachememory.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}
func (c *Cache) GetAll() (kv []*cachememory.Entity) {
	if c.cachememory == nil {
		return []*cachememory.Entity{}
	}

	return c.cachememory.GetAll()
}

func (c *Cache) TTL(key string) int64 {
	if c.cachememory == nil {
		return -2
	}
	return c.cachememory.TTL(key)
}
