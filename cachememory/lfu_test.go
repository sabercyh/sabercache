package cachememory

import (
	"reflect"
	"testing"
)

func TestLFUGet(t *testing.T) {
	var cache CacheMemory = NewLFUCache(int64(1024), nil)
	cache.Add("key1", String("1234"))
	if v, ok := cache.Get("key1"); !ok || v.(String) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := cache.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestLFURemove(t *testing.T) {
	k1, k2, k3, k4, k5, k6 := "key1", "key2", "key3", "key4", "key5", "key6key6key6"
	v1, v2, v3, v4, v5, v6 := "value1", "value2", "value3", "value4", "value5", "value6value6value6"
	cap := len(k1 + k2 + v1 + v2 + k3 + v3)
	var cache CacheMemory = NewLFUCache(int64(cap), nil)
	t.Run("r1", func(t *testing.T) {
		cache.Add(k1, String(v1))
		cache.Add(k2, String(v2))
		cache.Add(k3, String(v3))
		cache.Add(k4, String(v4))

		if _, ok := cache.Get(k1); ok || cache.Len() != 3 {
			t.Fatalf("Removeoldest key1 failed")
		}
	})
	t.Run("r2", func(t *testing.T) {
		cache.Add(k1, String(v1))
		cache.Add(k2, String(v2))
		cache.Add(k3, String(v3))
		cache.Add(k4, String(v4))
		// cache.Get(k2)
		cache.Add(k2, String("value"))
		cache.Add(k5, String(v5))
		if _, ok := cache.Get(k3); ok || cache.Len() != 3 {
			t.Fatalf("Removeoldest key3 failed")
		}
	})
	t.Run("r3", func(t *testing.T) {
		cache.Add(k1, String(v1))
		cache.Add(k2, String(v2))
		cache.Add(k3, String(v3))
		cache.Add(k4, String(v4))
		// cache.Get(k2)
		cache.Add(k2, String("value"))
		cache.Add(k5, String(v5))
		cache.Add(k6, String(v6))
		if _, ok := cache.Get(k5); ok || cache.Len() != 1 {
			t.Fatalf("Removeoldest key2,key4,key5 failed")
		}
	})

}

func TestLFUOnEnvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	var cache CacheMemory = NewLFUCache(int64(10), callback)
	cache.Add("key1", String("123456"))
	cache.Add("k2", String("k2"))
	cache.Add("k3", String("k3"))
	cache.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
