package cachememory

import "container/list"

type LRUCache struct {
	capacity         int64 // Cache 最大容量(Byte)
	length           int64 // Cache 当前容量(Byte)
	hashmap          map[string]*list.Element
	doublyLinkedList *list.List // 链头表示最近使用

	callback OnEliminated
}

func NewLRUCache(maxBytes int64, callback OnEliminated) *LRUCache {
	return &LRUCache{
		capacity:         maxBytes,
		hashmap:          make(map[string]*list.Element),
		doublyLinkedList: list.New(),
		callback:         callback,
	}
}

// Get 从缓存获取对应key的value。
// ok 指明查询结果 false代表查无此key
func (c *LRUCache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.hashmap[key]; ok {
		c.doublyLinkedList.MoveToFront(elem)
		entity := elem.Value.(*Entity)
		return entity.value, true
	}
	return
}

func (c *LRUCache) Add(key string, value Value) {
	kvSize := int64(len(key)) + int64(value.Len())
	// cache 容量检查
	for c.capacity != 0 && c.length+kvSize > c.capacity {
		c.Remove()
	}
	if elem, ok := c.hashmap[key]; ok {
		// 更新缓存key值
		c.doublyLinkedList.MoveToFront(elem)
		oldEntry := elem.Value.(*Entity)
		// 先更新写入字节 再更新
		c.length += int64(value.Len()) - int64(oldEntry.value.Len())
		oldEntry.value = value
	} else {
		// 新增缓存key
		elem := c.doublyLinkedList.PushFront(&Entity{key: key, value: value})
		c.hashmap[key] = elem
		c.length += kvSize
	}
}

// Remove 淘汰一枚最近最不常用缓存
func (c *LRUCache) Remove() {
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
func (c *LRUCache) Len() int {
	return c.doublyLinkedList.Len()
}

var _ CacheMemory = (*LRUCache)(nil)
