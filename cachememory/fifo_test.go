package cachememory

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestFIFOGet(t *testing.T) {
	var cache CacheMemory = NewFIFOCache(int64(1024), nil)
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
func TestFIFOSetWithTTL(t *testing.T) {
	var cache CacheMemory = NewFIFOCache(int64(1024), nil)
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
func TestFIFOMultiDeleteKey(t *testing.T) {
	var cache CacheMemory = NewFIFOCache(int64(1024), nil)
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
func TestFIFOTTL(t *testing.T) {
	var cache CacheMemory = NewFIFOCache(int64(1024), nil)
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
func TestFIFORemove(t *testing.T) {
	k1, k2, k3, k4, k5 := "key1", "key2", "key3", "key4key4", "key5"
	v1, v2, v3, v4, v5 := "value1", "value2", "value3", "value4value4", "value5"
	cap := len(k1 + k2 + v1 + v2)
	var cache CacheMemory = NewFIFOCache(int64(cap), nil)
	go cache.ExpireKeyMonitor()
	defer cache.Stop()
	cache.SetWithoutTTL(k1, String(v1))
	cache.SetWithoutTTL(k2, String(v2))
	cache.SetWithoutTTL(k3, String(v3))

	if _, ok := cache.Get(k1); ok || cache.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
	cache.SetWithoutTTL(k4, String(v4))
	if _, ok := cache.Get(k3); ok || cache.Len() != 1 {
		t.Fatalf("Removeoldest key2,key3 failed")
	}
	cache.SetWithoutTTL(k5, String(v5))
	if _, ok := cache.Get(k4); ok || cache.Len() != 1 {
		t.Fatalf("Removeoldest key4 failed")
	}

}
func TestFIFOMaxSize(t *testing.T) {
	t.Run("SetWithoutTTL", func(t *testing.T) {
		var cache CacheMemory = NewFIFOCache(int64(4), nil)
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
		var cache CacheMemory = NewFIFOCache(int64(4), nil)
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
func TestFIFOOnEnvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	var cache CacheMemory = NewFIFOCache(int64(10), callback)
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
