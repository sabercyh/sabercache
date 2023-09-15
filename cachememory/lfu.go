package cachememory

import (
	"container/list"
	"sync"
	"time"
)

type LFUCache struct {
	capacity         int64
	length           int64
	doublyLinkedList *list.List
	hashmap          map[string]*list.Element
	valuefreqmap     map[*list.Element]*ValueFreq
	freqmap          map[int]*list.List
	timemap          map[int64][]string
	mu               sync.Mutex
	stop             chan struct{}
	callback         OnEliminated
}

type ValueFreq struct {
	freq int
	elem *list.Element
}

func NewLFUCache(capacity int64, callback OnEliminated) *LFUCache {
	return &LFUCache{
		capacity:         capacity,
		doublyLinkedList: list.New(),
		valuefreqmap:     make(map[*list.Element]*ValueFreq),
		hashmap:          make(map[string]*list.Element),
		freqmap:          make(map[int]*list.List),
		timemap:          make(map[int64][]string),
		stop:             make(chan struct{}),
		callback:         callback,
	}
}

func (c *LFUCache) Get(key string) (Value, bool) {
	c.mu.Lock()
	if elem, ok := c.hashmap[key]; ok {
		entity := elem.Value.(*Entity)
		if entity.expiredTime != -1 && entity.expiredTime <= time.Now().Unix() {
			c.mu.Unlock()
			c.RemoveExpiredKey(key)
			return nil, false
		}
		var moved bool
		freq := c.valuefreqmap[elem].freq
		e := c.valuefreqmap[elem].elem
		freq++
		if _, ok := c.freqmap[freq]; ok {
			front := c.freqmap[freq].Back().Value.(*list.Element)
			c.doublyLinkedList.MoveBefore(elem, front)
			moved = true
			e = c.freqmap[freq].PushBack(elem)
		} else {
			c.freqmap[freq] = list.New()
			e = c.freqmap[freq].PushBack(elem)
		}
		if c.freqmap[c.valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
			delete(c.freqmap, c.valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
		}
		c.valuefreqmap[elem].freq++
		c.valuefreqmap[elem].elem = e
		c.mu.Unlock()
		return entity.value, true
	}
	c.mu.Unlock()
	return nil, false
}

func (c *LFUCache) SetWithoutTTL(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(key)) + int64(value.Len())
	if kvSize > c.capacity {
		return
	}
	if elem, ok := c.hashmap[key]; ok {
		var moved bool
		freq := c.valuefreqmap[elem].freq
		e := c.valuefreqmap[elem].elem
		freq++
		if _, ok := c.freqmap[freq]; ok {
			front := c.freqmap[freq].Back().Value.(*list.Element)
			c.doublyLinkedList.MoveBefore(elem, front)
			moved = true
			e = c.freqmap[freq].PushBack(elem)
		} else {
			c.freqmap[freq] = list.New()
			e = c.freqmap[freq].PushBack(elem)
		}
		if c.freqmap[c.valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
			delete(c.freqmap, c.valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
		}
		c.valuefreqmap[elem].freq++
		c.valuefreqmap[elem].elem = e
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.expiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length-int64(oldEntry.value.Len())+int64(value.Len()) > c.capacity {
			c.Remove()
		}
		c.length += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
		oldEntry.expiredTime = -1
	} else {
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		c.Push(&Entity{key: key, value: value, expiredTime: -1})
		c.length += kvSize
	}
}
func (c *LFUCache) SetWithTTL(key string, value Value, expireSecond int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(key)) + int64(value.Len())
	if kvSize > c.capacity {
		return
	}
	expireTime := time.Now().Unix() + expireSecond
	if elem, ok := c.hashmap[key]; ok {
		var moved bool
		freq := c.valuefreqmap[elem].freq
		e := c.valuefreqmap[elem].elem
		freq++
		if _, ok := c.freqmap[freq]; ok {
			front := c.freqmap[freq].Back().Value.(*list.Element)
			c.doublyLinkedList.MoveBefore(elem, front)
			moved = true
			e = c.freqmap[freq].PushBack(elem)
		} else {
			c.freqmap[freq] = list.New()
			e = c.freqmap[freq].PushBack(elem)
		}
		if c.freqmap[c.valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
			delete(c.freqmap, c.valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.valuefreqmap[elem].freq].Remove(c.valuefreqmap[elem].elem)
		}
		c.valuefreqmap[elem].freq++
		c.valuefreqmap[elem].elem = e
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.expiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length-int64(oldEntry.value.Len())+int64(value.Len()) > c.capacity {
			c.Remove()
		}
		c.length += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
		oldEntry.expiredTime = expireTime
	} else {
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		c.Push(&Entity{key: key, value: value, expiredTime: expireTime})
		c.length += kvSize
	}
	c.timemap[expireTime] = append(c.timemap[expireTime], key)
}

func (c *LFUCache) Push(entity *Entity) {
	if _, ok := c.freqmap[1]; !ok {
		elem := c.doublyLinkedList.PushBack(entity)
		c.hashmap[entity.key] = elem
		c.freqmap[1] = list.New()
		e := c.freqmap[1].PushBack(elem)
		c.valuefreqmap[elem] = &ValueFreq{1, e}
	} else {
		front := c.freqmap[1].Back().Value.(*list.Element)
		elem := c.doublyLinkedList.InsertBefore(entity, front)
		c.hashmap[entity.key] = elem
		e := c.freqmap[1].PushBack(elem)
		c.valuefreqmap[elem] = &ValueFreq{1, e}
	}
}
func (c *LFUCache) ExpireKeyMonitor() {
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
func (c *LFUCache) RemoveExpiredKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.hashmap[key]; ok {
		freq := c.valuefreqmap[elem].freq
		e := c.valuefreqmap[elem].elem
		delete(c.valuefreqmap, elem)
		c.freqmap[freq].Remove(e)
		if c.freqmap[freq].Front() == nil {
			delete(c.freqmap, freq)
		}
		key := elem.Value.(*Entity).key
		value := elem.Value.(*Entity).value
		delete(c.hashmap, key)
		c.doublyLinkedList.Remove(elem)
		c.length = c.length - int64(len(key)) - int64(value.Len())

		if c.callback != nil {
			c.callback(key, value)
		}
	} else {
		return
	}
}
func (c *LFUCache) MultiDeleteKey(keys []string, t int64) {
	c.mu.Lock()
	delete(c.timemap, t)
	c.mu.Unlock()
	for _, v := range keys {
		c.RemoveExpiredKey(v)
	}
}

func (c *LFUCache) Remove() {
	elem := c.doublyLinkedList.Back()
	freq := c.valuefreqmap[elem].freq
	e := c.valuefreqmap[elem].elem
	delete(c.valuefreqmap, elem)
	c.freqmap[freq].Remove(e)
	if c.freqmap[freq].Front() == nil {
		delete(c.freqmap, freq)
	}
	key := elem.Value.(*Entity).key
	value := elem.Value.(*Entity).value
	delete(c.hashmap, key)
	c.doublyLinkedList.Remove(elem)
	c.length = c.length - int64(len(key)) - int64(value.Len())

	if c.callback != nil {
		c.callback(key, value)
	}
}
func (c *LFUCache) Len() int {
	return c.doublyLinkedList.Len()
}
func (c *LFUCache) TTL(key string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
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
func (c *LFUCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stop <- struct{}{}
}
func (c *LFUCache) Stop() {
	c.Close()
}

var _ CacheMemory = (*LFUCache)(nil)
