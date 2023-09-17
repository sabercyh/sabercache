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
	Valuefreqmap     map[*list.Element]*ValueFreq
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
		Valuefreqmap:     make(map[*list.Element]*ValueFreq),
		hashmap:          make(map[string]*list.Element),
		freqmap:          make(map[int]*list.List),
		timemap:          make(map[int64][]string),
		stop:             make(chan struct{}),
		callback:         callback,
	}
}

func (c *LFUCache) Get(Key string) (Value, bool) {
	c.mu.Lock()
	if elem, ok := c.hashmap[Key]; ok {
		entity := elem.Value.(*Entity)
		if entity.ExpiredTime != -1 && entity.ExpiredTime <= time.Now().Unix() {
			c.mu.Unlock()
			c.RemoveExpiredKey(Key)
			return nil, false
		}
		var moved bool
		freq := c.Valuefreqmap[elem].freq
		e := c.Valuefreqmap[elem].elem
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
		if c.freqmap[c.Valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
			delete(c.freqmap, c.Valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.Valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
		}
		c.Valuefreqmap[elem].freq++
		c.Valuefreqmap[elem].elem = e
		c.mu.Unlock()
		return entity.Value, true
	}
	c.mu.Unlock()
	return nil, false
}
func (c *LFUCache) GetAll() (kv []*Entity) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, elem := range c.hashmap {
		entity := elem.Value.(*Entity)
		if entity.ExpiredTime != -1 && entity.ExpiredTime <= time.Now().Unix() {
			continue
		}
		kv = append(kv, &Entity{Key: entity.Key, Value: entity.Value})
	}
	return
}
func (c *LFUCache) SetWithoutTTL(Key string, Value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(Key)) + int64(Value.Len())
	if kvSize > c.capacity {
		return
	}
	if elem, ok := c.hashmap[Key]; ok {
		var moved bool
		freq := c.Valuefreqmap[elem].freq
		e := c.Valuefreqmap[elem].elem
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
		if c.freqmap[c.Valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
			delete(c.freqmap, c.Valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.Valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
		}
		c.Valuefreqmap[elem].freq++
		c.Valuefreqmap[elem].elem = e
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.ExpiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.Key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length-int64(oldEntry.Value.Len())+int64(Value.Len()) > c.capacity {
			c.Remove()
		}
		c.length += int64(Value.Len()) - int64(oldEntry.Value.Len())
		oldEntry.Value = Value
		oldEntry.ExpiredTime = -1
	} else {
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		c.Push(&Entity{Key: Key, Value: Value, ExpiredTime: -1})
		c.length += kvSize
	}
}
func (c *LFUCache) SetWithTTL(Key string, Value Value, ttl int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	kvSize := int64(len(Key)) + int64(Value.Len())
	if kvSize > c.capacity {
		return
	}
	expireTime := time.Now().Unix() + ttl
	if elem, ok := c.hashmap[Key]; ok {
		var moved bool
		freq := c.Valuefreqmap[elem].freq
		e := c.Valuefreqmap[elem].elem
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
		if c.freqmap[c.Valuefreqmap[elem].freq].Len() == 1 {
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
			delete(c.freqmap, c.Valuefreqmap[elem].freq)
		} else {
			if !moved {
				front := c.freqmap[c.Valuefreqmap[elem].freq].Back().Value.(*list.Element)
				c.doublyLinkedList.MoveBefore(elem, front)
			}
			c.freqmap[c.Valuefreqmap[elem].freq].Remove(c.Valuefreqmap[elem].elem)
		}
		c.Valuefreqmap[elem].freq++
		c.Valuefreqmap[elem].elem = e
		oldEntry := elem.Value.(*Entity)
		if strs, ok := c.timemap[oldEntry.ExpiredTime]; ok {
			for i, v := range strs {
				if v == oldEntry.Key {
					strs[i] = ""
					break
				}
			}
		}
		for c.capacity != 0 && c.length-int64(oldEntry.Value.Len())+int64(Value.Len()) > c.capacity {
			c.Remove()
		}
		c.length += int64(Value.Len()) - int64(oldEntry.Value.Len())
		oldEntry.Value = Value
		oldEntry.ExpiredTime = expireTime
	} else {
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		c.Push(&Entity{Key: Key, Value: Value, ExpiredTime: expireTime})
		c.length += kvSize
	}
	c.timemap[expireTime] = append(c.timemap[expireTime], Key)
}

func (c *LFUCache) Push(entity *Entity) {
	if _, ok := c.freqmap[1]; !ok {
		elem := c.doublyLinkedList.PushBack(entity)
		c.hashmap[entity.Key] = elem
		c.freqmap[1] = list.New()
		e := c.freqmap[1].PushBack(elem)
		c.Valuefreqmap[elem] = &ValueFreq{1, e}
	} else {
		front := c.freqmap[1].Back().Value.(*list.Element)
		elem := c.doublyLinkedList.InsertBefore(entity, front)
		c.hashmap[entity.Key] = elem
		e := c.freqmap[1].PushBack(elem)
		c.Valuefreqmap[elem] = &ValueFreq{1, e}
	}
}
func (c *LFUCache) ExpireKeyMonitor() {
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
func (c *LFUCache) RemoveExpiredKey(Key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.hashmap[Key]; ok {
		freq := c.Valuefreqmap[elem].freq
		e := c.Valuefreqmap[elem].elem
		delete(c.Valuefreqmap, elem)
		c.freqmap[freq].Remove(e)
		if c.freqmap[freq].Front() == nil {
			delete(c.freqmap, freq)
		}
		Key := elem.Value.(*Entity).Key
		Value := elem.Value.(*Entity).Value
		delete(c.hashmap, Key)
		c.doublyLinkedList.Remove(elem)
		c.length = c.length - int64(len(Key)) - int64(Value.Len())

		if c.callback != nil {
			c.callback(Key, Value)
		}
	} else {
		return
	}
}
func (c *LFUCache) MultiDeleteKey(Keys []string, t int64) {
	c.mu.Lock()
	delete(c.timemap, t)
	c.mu.Unlock()
	for _, v := range Keys {
		c.RemoveExpiredKey(v)
	}
}

func (c *LFUCache) Remove() {
	elem := c.doublyLinkedList.Back()
	freq := c.Valuefreqmap[elem].freq
	e := c.Valuefreqmap[elem].elem
	delete(c.Valuefreqmap, elem)
	c.freqmap[freq].Remove(e)
	if c.freqmap[freq].Front() == nil {
		delete(c.freqmap, freq)
	}
	Key := elem.Value.(*Entity).Key
	Value := elem.Value.(*Entity).Value
	delete(c.hashmap, Key)
	c.doublyLinkedList.Remove(elem)
	c.length = c.length - int64(len(Key)) - int64(Value.Len())

	if c.callback != nil {
		c.callback(Key, Value)
	}
}
func (c *LFUCache) Len() int {
	return c.doublyLinkedList.Len()
}
func (c *LFUCache) TTL(Key string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
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
func (c *LFUCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stop <- struct{}{}
}
func (c *LFUCache) Stop() {
	c.Close()
}

var _ CacheMemory = (*LFUCache)(nil)
