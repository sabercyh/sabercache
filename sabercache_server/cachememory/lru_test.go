package cachememory

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestLRU(t *testing.T) {
	k1, k2, k3, k4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cap := len(k1 + k2 + k3 + v1 + v2 + v3)

	t.Run("过期后淘汰", func(t *testing.T) {
		var cache CacheMemory = NewLRUCache(int64(cap), nil)
		go cache.ExpireKeyMonitor()
		defer cache.Stop()
		cache.SetWithoutTTL(k1, String(v1))
		cache.SetWithTTL(k2, String(v2), 3)
		cache.SetWithTTL(k3, String(v3), 3)
		time.Sleep(4 * time.Second)
		cache.SetWithoutTTL(k4, String(v4))
		if _, ok := cache.Get(k1); !ok {
			t.Fatalf("Get key1 fialed!")
		}
		if _, ok := cache.Get(k4); !ok {
			t.Fatalf("Get key4 fialed!")
		}
	})
	t.Run("查询刷新", func(t *testing.T) {
		var cache CacheMemory = NewLRUCache(int64(cap), nil)
		go cache.ExpireKeyMonitor()
		defer cache.Stop()
		cache.SetWithoutTTL(k1, String(v1))
		cache.SetWithTTL(k2, String(v2), 3)
		cache.SetWithoutTTL(k3, String(v3))
		time.Sleep(1 * time.Second)
		cache.SetWithoutTTL(k2, String(v2))
		time.Sleep(2 * time.Second)
		if _, ok := cache.Get(k2); !ok {
			t.Fatalf("Get key2 fialed!")
		}
		cache.SetWithoutTTL(k4, String(v4))
		if _, ok := cache.Get(k1); ok {
			t.Fatalf("Remove key3 fialed!")
		}
		if _, ok := cache.Get(k3); !ok {
			t.Fatalf("Get key3 fialed!")
		}

	})

}
func TestLRUGet(t *testing.T) {
	var cache CacheMemory = NewLRUCache(int64(1024), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	t.Run("SetWithoutTTL", func(t *testing.T) {
		cache.SetWithoutTTL("key1", String("1234"))
		if v, ok := cache.Get("key1"); !ok || v.(String) != "1234" {
			t.Fatalf("cache hit key1=1234 failed")
		}
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		cache.SetWithTTL("key1", String("1234"), 10)
		if v, ok := cache.Get("key1"); !ok || v.(String) != "1234" {
			t.Fatalf("cache hit key1=1234 failed")
		}
	})
	t.Run("GetNilKey", func(t *testing.T) {
		if _, ok := cache.Get("key2"); ok {
			t.Fatalf("cache miss key2 failed")
		}
	})
}
func TestLRUGetAll(t *testing.T) {
	var cache CacheMemory = NewLRUCache(int64(1024), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	t.Run("SetWithoutTTL", func(t *testing.T) {
		cache.SetWithoutTTL("key1", String("value1"))
		cache.SetWithoutTTL("key2", String("value2"))
		entity := cache.GetAll()
		if len(entity) != 2 {
			t.Fatalf("getall failed!")
		}
		fmt.Println(*entity[0], *entity[1])
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		cache.SetWithTTL("key1", String("value1"), 10)
		cache.SetWithTTL("key2", String("value2"), 10)
		entity := cache.GetAll()
		if len(entity) != 2 {
			t.Fatalf("getall failed!")
		}
		fmt.Println(*entity[0], *entity[1])
	})
}
func TestLRUSetWithTTL(t *testing.T) {
	var cache CacheMemory = NewLRUCache(int64(1024), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	t.Run("expireTest", func(t *testing.T) {
		cache.SetWithTTL("key1", String("1234"), 5)
		time.Sleep(5 * time.Second)
		if _, ok := cache.Get("key1"); ok {
			t.Fatalf("cache remove expired Key key1 failed")
		}
	})
}
func TestLRUMultiDeleteKey(t *testing.T) {
	var cache CacheMemory = NewLRUCache(int64(1024), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	now := time.Now().Unix() + 5
	keys := []string{"key1", "key2", "key3"}
	go cache.SetWithTTL("key1", String("value1"), 5)
	go cache.SetWithTTL("key2", String("value2"), 5)
	go cache.SetWithTTL("key3", String("value3"), 5)
	time.Sleep(3 * time.Second)
	for i := 1; i <= 3; i++ {
		if _, ok := cache.Get("key" + strconv.Itoa(i)); !ok {
			t.Fatalf("cache hit key" + strconv.Itoa(i) + " failed")
		}
	}
	cache.MultiDeleteKey(keys, now)
	for i := 1; i <= 3; i++ {
		if _, ok := cache.Get("key" + strconv.Itoa(i)); ok {
			t.Fatalf("multiple delete failed")
		}
	}

}
func TestLRUTTL(t *testing.T) {
	var cache CacheMemory = NewLRUCache(int64(1024), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	t.Run("TTLTest", func(t *testing.T) {
		cache.SetWithTTL("key1", String("1234"), 5)
		if ttl := cache.TTL("key1"); ttl != 5 {
			t.Fatalf("ttl test key1 failed")
		}
	})
	t.Run("TTLExpireTest", func(t *testing.T) {
		cache.SetWithTTL("key1", String("1234"), 2)
		time.Sleep(2 * time.Second)
		if ttl := cache.TTL("key1"); ttl != -2 {
			t.Fatalf("ttl test expiredKey key1 failed")
		}
	})
}
func TestLRURemove(t *testing.T) {
	k1, k2, k3, k4, k5 := "key1", "key2", "key3", "key4", "key5key5"
	v1, v2, v3, v4, v5 := "value1", "value2", "value3", "value4", "value5value5"
	cap := len(k1 + k2 + v1 + v2)
	var cache CacheMemory = NewLRUCache(int64(cap), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	t.Run("SetWithoutTTL", func(t *testing.T) {
		cache.SetWithoutTTL(k1, String(v1))
		cache.SetWithoutTTL(k2, String(v2))
		cache.SetWithoutTTL(k3, String(v3))

		if _, ok := cache.Get(k1); ok || cache.Len() != 2 {
			t.Fatalf("Removeoldest key1 failed")
		}

		cache.SetWithoutTTL(k2, String("value2"))
		cache.SetWithoutTTL(k4, String(v4))
		if _, ok := cache.Get(k3); ok || cache.Len() != 2 {
			t.Fatalf("Removeoldest key3 failed")
		}
		//len(value5)==cap
		cache.SetWithoutTTL(k5, String(v5))
		if _, ok := cache.Get(k4); ok || cache.Len() != 1 {
			t.Fatalf("Removeoldest key4 failed")
		}
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		cache.SetWithTTL(k1, String(v1), 10)
		cache.SetWithTTL(k2, String(v2), 10)
		cache.SetWithTTL(k3, String(v3), 10)

		if _, ok := cache.Get(k1); ok || cache.Len() != 2 {
			t.Fatalf("Removeoldest key1 failed")
		}

		cache.SetWithTTL(k2, String("value2"), 10)
		cache.SetWithTTL(k4, String(v4), 10)
		if _, ok := cache.Get(k3); ok || cache.Len() != 2 {
			t.Fatalf("Removeoldest key3 failed")
		}
		//len(value5)==cap
		cache.SetWithTTL(k5, String(v5), 10)
		if _, ok := cache.Get(k4); ok || cache.Len() != 1 {
			t.Fatalf("Removeoldest key4 failed")
		}
	})

}
func TestLRUMaxSize(t *testing.T) {
	t.Run("SetWithoutTTL", func(t *testing.T) {
		var cache CacheMemory = NewLRUCache(int64(4), nil)
		go cache.ExpireKeyMonitor()
		defer cache.Stop()
		cache.SetWithoutTTL("key1", String("value1"))
		if _, ok := cache.Get("key1"); ok {
			t.Fatalf("Get out of size Key key1")
		}
		cache.SetWithoutTTL("k2", String("v2"))
		if _, ok := cache.Get("k2"); !ok {
			t.Fatalf("Get k2 fialed")
		}
		cache.SetWithoutTTL("k2", String("value2"))
		if v2, ok := cache.Get("k2"); !ok || v2 != String("v2") {
			t.Fatalf("Update out of size Key k2")
		}
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		var cache CacheMemory = NewLRUCache(int64(4), nil)
		go cache.ExpireKeyMonitor()
		defer cache.Stop()
		cache.SetWithTTL("key1", String("value1"), 10)
		if _, ok := cache.Get("key1"); ok {
			t.Fatalf("Get out of size Key key1")
		}
		cache.SetWithTTL("k2", String("v2"), 10)
		if _, ok := cache.Get("k2"); !ok {
			t.Fatalf("Get k2 fialed")
		}
		cache.SetWithTTL("k2", String("value2"), 10)
		if v2, ok := cache.Get("k2"); !ok || v2 != String("v2") {
			t.Fatalf("Update out of size Key k2")
		}
	})
}
func TestLRUOnEnvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	var cache CacheMemory = NewLRUCache(int64(10), callback)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	cache.SetWithoutTTL("key1", String("123456"))
	cache.SetWithoutTTL("k2", String("k2"))
	cache.SetWithoutTTL("k3", String("k3"))
	cache.SetWithoutTTL("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
