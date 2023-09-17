package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func(data []byte) uint32

type Consistency struct {
	hash     HashFunc
	replicas int
	ring     []int
	hashmap  map[int]string
}

func (c *Consistency) Register(peersName []string) {
	for _, peerName := range peersName {
		for i := 0; i < c.replicas; i++ {
			hashValue := int(c.hash([]byte(strconv.Itoa(i) + peerName)))
			c.ring = append(c.ring, hashValue)
			c.hashmap[hashValue] = peerName
		}
	}
	sort.Ints(c.ring)
}

func (c *Consistency) GetPeer(key string) string {
	if len(c.ring) == 0 {
		return ""
	}
	hashValue := int(c.hash([]byte(key)))
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hashValue
	})
	return c.hashmap[c.ring[idx%len(c.ring)]]
}

func New(replicas int, fn HashFunc) *Consistency {
	c := &Consistency{
		replicas: replicas,
		hash:     fn,
		hashmap:  make(map[int]string),
	}
	if c.hash == nil {
		c.hash = crc32.ChecksumIEEE
	}
	return c
}
