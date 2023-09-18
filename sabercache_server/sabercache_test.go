package sabercache_server

import (
	"fmt"
	"log"
	"sabercache_server/cachememory"
	"testing"
)

func TestGet(t *testing.T) {
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	sc := NewSaberCache(2<<10, "fifo", 30, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	t.Run("Get", func(t *testing.T) {
		sc.Set("k1", ByteView{[]byte("v1")}, -2)
		if v, err := sc.Get("k1"); err != nil || v.String() != "v1" {
			t.Fatalf("get k1 fialed")
		}
	})
	t.Run("Load", func(t *testing.T) {
		if v, err := sc.Get("Tom"); err != nil || v.String() != "630" {
			t.Fatalf("load Tom fialed")
		}
	})
	t.Run("getnil", func(t *testing.T) {
		if _, err := sc.Get("AA"); err == nil {
			t.Fatalf("load nil key!")
		}
	})
}
func TestGetAll(t *testing.T) {
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	sc := NewSaberCache(2<<10, "fifo", 30, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	t.Run("GetAll", func(t *testing.T) {
		var entity []*cachememory.Entity
		sc.Set("k1", ByteView{[]byte("v1")}, -2)
		sc.Set("k2", ByteView{[]byte("v2")}, -2)
		if entity = sc.GetAll(); len(entity) != 2 {
			t.Fatalf("getall fialed")
		}
		fmt.Println(*entity[0], *entity[1])
	})
}

func TestSet(t *testing.T) {
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	sc := NewSaberCache(2<<10, "fifo", 30, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	t.Run("SetWithoutTTL", func(t *testing.T) {
		sc.Set("k1", ByteView{[]byte("v1")}, -2)
		if v, err := sc.Get("k1"); err != nil || v.String() != "v1" {
			t.Fatalf("get k1 fialed")
		}
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		sc.Set("k2", ByteView{[]byte("v2")}, 10)
		if v, err := sc.Get("k2"); err != nil || v.String() != "v2" {
			t.Fatalf("get k2 fialed")
		}
	})
}

func TestTTL(t *testing.T) {
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	sc := NewSaberCache(2<<10, "fifo", 30, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	t.Run("SetWithoutTTL", func(t *testing.T) {
		sc.Set("k1", ByteView{[]byte("v1")}, -1)
		if ttl := sc.TTL("k1"); ttl != -1 {
			t.Fatalf("set k1 without ttl fialed")
		}
	})
	t.Run("SetWithTTL", func(t *testing.T) {
		sc.Set("k2", ByteView{[]byte("v2")}, 10)
		if ttl := sc.TTL("k2"); ttl != 10 && ttl != 9 {
			t.Fatalf("set k2 with ttl fialed")
		}
	})
}
