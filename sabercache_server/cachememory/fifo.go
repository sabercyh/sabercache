package cachememory

import (
	"container/list"
	"sync"
	"time"
)

type FIFOCache struct {
	capacity         int64 // Cache 最大容量(Byte)
	length           int64 // Cache 当前容量(Byte)
	hashmap          map[string]*list.Element
	timemap          map[int64][]string
	doublyLinkedList *list.List // 链头表示最近使用
	mu               sync.RWMutex
	stop             chan struct{}
	callback         OnEliminated
}

func NewFIFOCache(maxBytes int64, callback OnEliminated) *FIFOCache {
	return &FIFOCache{
		capacity:         maxBytes,
		hashmap:          make(map[string]*list.Element),
		timemap:          make(map[int64][]string),
		doublyLinkedList: list.New(),
		callback:         callback,
		stop:             make(chan struct{}),
	}
}

// Get 从缓存获取对应Key的Value。
// ok 指明查询结果 false代表查无此Key
func (c *FIFOCache) Get(Key string) (Value Value, ok bool) {
	c.mu.RLock()
	if elem, ok := c.hashmap[Key]; ok {
		entity := elem.Value.(*Entity)
		c.mu.RUnlock()
		if entity.ExpiredTime != -1 && entity.ExpiredTime <= time.Now().Unix() {
			c.RemoveExpiredKey(Key)
			return nil, false
		}
		return entity.Value, true
	}
	c.mu.RUnlock()
	return
}
func (c *FIFOCache) GetAll() (kv []*Entity) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, elem := range c.hashmap {
		entity := elem.Value.(*Entity)
		if entity.ExpiredTime != -1 && entity.ExpiredTime <= time.Now().Unix() {
			continue
		}
		kv = append(kv, &Entity{Key: entity.Key, Value: entity.Value})
	}
	return
}

func (c *FIFOCache) SetWithoutTTL(Key string, Value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(Key)) + int64(Value.Len())
	if kvSize > c.capacity {
		return
	}
	if elem, ok := c.hashmap[Key]; ok {
		// 更新缓存Key值
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.ExpiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.Key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length+int64(Value.Len())-int64(oldEntry.Value.Len()) > c.capacity {
			c.Remove()
		}
		// 先更新写入字节 再更新
		c.length += int64(Value.Len()) - int64(oldEntry.Value.Len())
		oldEntry.Value = Value
		oldEntry.ExpiredTime = -1
	} else {
		// 新增缓存Key
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		elem := c.doublyLinkedList.PushFront(&Entity{Key: Key, Value: Value, ExpiredTime: -1})
		c.hashmap[Key] = elem
		c.length += kvSize
	}
}

func (c *FIFOCache) SetWithTTL(Key string, Value Value, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(Key)) + int64(Value.Len())
	if kvSize > c.capacity {
		return
	}
	expireTime := time.Now().Unix() + ttl
	if elem, ok := c.hashmap[Key]; ok {
		// 更新缓存Key值
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.ExpiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.Key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length+int64(Value.Len())-int64(oldEntry.Value.Len()) > c.capacity {
			c.Remove()
		}
		// 先更新写入字节 再更新
		c.length += int64(Value.Len()) - int64(oldEntry.Value.Len())
		oldEntry.Value = Value
		oldEntry.ExpiredTime = expireTime
	} else {
		// 新增缓存Key
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		elem := c.doublyLinkedList.PushFront(&Entity{Key: Key, Value: Value, ExpiredTime: expireTime})
		c.hashmap[Key] = elem
		c.length += kvSize
	}
	c.timemap[expireTime] = append(c.timemap[expireTime], Key)
}
func (c *FIFOCache) ExpireKeyMonitor() {
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	del := make(chan *DelCH, DelChCap)
	go func() {
		for v := range del {
			c.MultiDeleteKey(v.Keys, v.t)
		}
	}()
	now := time.Now().Unix()
	for {
		select {
		case <-t.C:
			now++
			c.mu.Lock()
			// fmt.Println(c.timemap)
			if Keys, ok := c.timemap[now]; ok {
				c.mu.Unlock()
				del <- &DelCH{Keys: Keys, t: now}
			} else {
				c.mu.Unlock()
			}
		case <-c.stop:
			close(del)
			return
		}
	}
}
func (c *FIFOCache) MultiDeleteKey(Keys []string, t int64) {
	c.mu.Lock()
	delete(c.timemap, t)
	c.mu.Unlock()
	for _, v := range Keys {
		c.RemoveExpiredKey(v)
	}
}

// Remove 按插入顺序淘汰缓存
func (c *FIFOCache) Remove() {
	tailElem := c.doublyLinkedList.Back()
	if tailElem != nil {
		entry := tailElem.Value.(*Entity)
		k, v := entry.Key, entry.Value
		delete(c.hashmap, k)                       // 移除映射
		c.doublyLinkedList.Remove(tailElem)        // 移除缓存
		c.length -= int64(len(k)) + int64(v.Len()) // 更新占用内存情况
		// 移除后的善后处理
		if c.callback != nil {
			c.callback(k, v)
		}
	}
}
func (c *FIFOCache) RemoveExpiredKey(Key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.hashmap[Key]; ok {
		entry := elem.Value.(*Entity)
		k, v := entry.Key, entry.Value
		delete(c.hashmap, k)                       // 移除映射
		c.doublyLinkedList.Remove(elem)            // 移除缓存
		c.length -= int64(len(k)) + int64(v.Len()) // 更新占用内存情况
		// 移除后的善后处理
		if c.callback != nil {
			c.callback(k, v)
		}
	} else {
		return
	}
}
func (c *FIFOCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.doublyLinkedList.Len()
}
func (c *FIFOCache) TTL(Key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if elem, ok := c.hashmap[Key]; ok {
		expireTime := elem.Value.(*Entity).ExpiredTime
		ttl := expireTime - time.Now().Unix()
		if expireTime == -1 {
			return -1
		} else {
			if ttl == 0 {
				return -2
			} else {
				return ttl
			}
		}
	} else {
		return -2
	}
}
func (c *FIFOCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stop <- struct{}{}
}
func (c *FIFOCache) Stop() {
	c.Close()
}

var _ CacheMemory = (*FIFOCache)(nil)
