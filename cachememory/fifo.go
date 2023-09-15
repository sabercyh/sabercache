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

// Get 从缓存获取对应key的value。
// ok 指明查询结果 false代表查无此key
func (c *FIFOCache) Get(key string) (value Value, ok bool) {
	c.mu.RLock()
	if elem, ok := c.hashmap[key]; ok {
		entity := elem.Value.(*Entity)
		c.mu.RUnlock()
		if entity.expiredTime != -1 && entity.expiredTime <= time.Now().Unix() {
			c.RemoveExpiredKey(key)
			return nil, false
		}
		return entity.value, true
	}
	c.mu.RUnlock()
	return
}

func (c *FIFOCache) SetWithoutTTL(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(key)) + int64(value.Len())
	if kvSize > c.capacity {
		return
	}
	if elem, ok := c.hashmap[key]; ok {
		// 更新缓存key值
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.expiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length+int64(value.Len())-int64(oldEntry.value.Len()) > c.capacity {
			c.Remove()
		}
		// 先更新写入字节 再更新
		c.length += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
		oldEntry.expiredTime = -1
	} else {
		// 新增缓存key
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		elem := c.doublyLinkedList.PushFront(&Entity{key: key, value: value, expiredTime: -1})
		c.hashmap[key] = elem
		c.length += kvSize
	}
}

func (c *FIFOCache) SetWithTTL(key string, value Value, expireSecond int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(key)) + int64(value.Len())
	if kvSize > c.capacity {
		return
	}
	expireTime := time.Now().Unix() + expireSecond
	if elem, ok := c.hashmap[key]; ok {
		// 更新缓存key值
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.expiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length+int64(value.Len())-int64(oldEntry.value.Len()) > c.capacity {
			c.Remove()
		}
		// 先更新写入字节 再更新
		c.length += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
		oldEntry.expiredTime = expireTime
	} else {
		// 新增缓存key
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		elem := c.doublyLinkedList.PushFront(&Entity{key: key, value: value, expiredTime: expireTime})
		c.hashmap[key] = elem
		c.length += kvSize
	}
	c.timemap[expireTime] = append(c.timemap[expireTime], key)
}
func (c *FIFOCache) ExpireKeyMonitor() {
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	del := make(chan *DelCH, DelChCap)
	go func() {
		for v := range del {
			c.MultiDeleteKey(v.keys, v.t)
		}
	}()
	now := time.Now().Unix()
	for {
		select {
		case <-t.C:
			now++
			c.mu.Lock()
			// fmt.Println(c.timemap)
			if keys, ok := c.timemap[now]; ok {
				c.mu.Unlock()
				del <- &DelCH{keys: keys, t: now}
			} else {
				c.mu.Unlock()
			}
		case <-c.stop:
			close(del)
			return
		}
	}
}
func (c *FIFOCache) MultiDeleteKey(keys []string, t int64) {
	c.mu.Lock()
	delete(c.timemap, t)
	c.mu.Unlock()
	for _, v := range keys {
		c.RemoveExpiredKey(v)
	}
}

// Remove 按插入顺序淘汰缓存
func (c *FIFOCache) Remove() {
	tailElem := c.doublyLinkedList.Back()
	if tailElem != nil {
		entry := tailElem.Value.(*Entity)
		k, v := entry.key, entry.value
		delete(c.hashmap, k)                       // 移除映射
		c.doublyLinkedList.Remove(tailElem)        // 移除缓存
		c.length -= int64(len(k)) + int64(v.Len()) // 更新占用内存情况
		// 移除后的善后处理
		if c.callback != nil {
			c.callback(k, v)
		}
	}
}
func (c *FIFOCache) RemoveExpiredKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.hashmap[key]; ok {
		entry := elem.Value.(*Entity)
		k, v := entry.key, entry.value
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
func (c *FIFOCache) TTL(key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if elem, ok := c.hashmap[key]; ok {
		expireTime := elem.Value.(*Entity).expiredTime
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
