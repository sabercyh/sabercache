package cachememory

import (
	"container/list"
)

type LFUCache struct {
	capacity         int64
	length           int64
	doublyLinkedList *list.List
	hashmap          map[string]*list.Element
	valuefreqmap     map[*list.Element]*ValueFreq
	freqmap          map[int]*list.List

	callback OnEliminated
}
type KeyValue struct {
	key   string
	value Value
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
		callback:         callback,
	}
}

func (c *LFUCache) Get(key string) (Value, bool) {
	if elem, ok := c.hashmap[key]; ok {
		var moved bool
		value := elem.Value.(*KeyValue).value
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
		return value, true
	} else {
		return nil, false
	}
}

func (c *LFUCache) Add(key string, value Value) {
	kvSize := int64(len(key)) + int64(value.Len())
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
		kv := elem.Value.(*KeyValue)
		for c.capacity != 0 && c.length-int64(kv.value.Len())+int64(value.Len()) > c.capacity {
			c.Remove()
		}
		c.length = c.length - int64(kv.value.Len()) + int64(value.Len())
		kv.value = value
	} else {
		for c.capacity != 0 && c.length+kvSize > c.capacity {
			c.Remove()
		}
		c.Push(&KeyValue{key, value})
		c.length = c.length + kvSize
	}
}
func (c *LFUCache) Push(kv *KeyValue) {
	if _, ok := c.freqmap[1]; !ok {
		elem := c.doublyLinkedList.PushBack(kv)
		c.hashmap[kv.key] = elem
		c.freqmap[1] = list.New()
		e := c.freqmap[1].PushBack(elem)
		c.valuefreqmap[elem] = &ValueFreq{1, e}
	} else {
		front := c.freqmap[1].Back().Value.(*list.Element)
		elem := c.doublyLinkedList.InsertBefore(kv, front)
		c.hashmap[kv.key] = elem
		e := c.freqmap[1].PushBack(elem)
		c.valuefreqmap[elem] = &ValueFreq{1, e}
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
	key := elem.Value.(*KeyValue).key
	value := elem.Value.(*KeyValue).value
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

var _ CacheMemory = (*LFUCache)(nil)
