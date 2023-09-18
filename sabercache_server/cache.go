package sabercache_server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sabercache_server/cachememory"
	"strconv"
	"strings"
	"time"
)

type Cache struct {
	cachememory   cachememory.CacheMemory
	capacity      int64
	cacheStrategy string
}

func newCache(capacity int64, cacheStrategy string) *Cache {
	c := &Cache{
		capacity:      capacity,
		cacheStrategy: cacheStrategy,
	}
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
	return c
}
func (c *Cache) Init() bool {
	file, error := os.Open("./backup/backup.txt")
	defer file.Close()
	if error != nil {
		log.Println(error)
		return false
	}
	now := time.Now().Unix()
	reader := bufio.NewReader(file)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Println(err)
				return false
			}
		}
		str = str[:len(str)-1]
		strs := strings.Split(str, " ")
		expireTime, err := strconv.Atoi(strs[2])
		if err != nil {
			log.Println(err)
			return false
		}
		if expireTime == -1 {
			c.SetWithoutTTL(strs[0], ByteView{[]byte(strs[1])})
		} else if int64(expireTime) > now {
			c.SetWithTTL(strs[0], ByteView{[]byte(strs[1])}, int64(expireTime))
		}
	}
	return true
}
func (c *Cache) SetWithoutTTL(key string, value ByteView) {
	c.cachememory.SetWithoutTTL(key, value)
}

func (c *Cache) SetWithTTL(key string, value ByteView, ttl int64) {
	c.cachememory.SetWithTTL(key, value, ttl)
}
func (c *Cache) Get(key string) (ByteView, bool) {
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

func (c *Cache) Save() bool {
	entitys := c.cachememory.GetAll()
	file, error := os.OpenFile("./backup/backup.txt", os.O_RDWR|os.O_CREATE, 0766)
	defer file.Close()
	if error != nil {
		log.Println(error)
		return false
	}
	writer := bufio.NewWriter(file)
	now := time.Now().Unix()
	for _, kv := range entitys {
		if kv.ExpiredTime == -1 || kv.ExpiredTime-now >= 30 {
			writer.WriteString(fmt.Sprintf("%s %s %d\n", kv.Key, kv.Value.(ByteView).String(), kv.ExpiredTime))
			writer.Flush()
		}
	}
	return true
}
