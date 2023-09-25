package sabercache_server

import (
	"fmt"
	"log"
	"math/rand"
	"sabercache_server/cachememory"
	"sabercache_server/singleflight"
)

var sabercache *SaberCache

type Retriever interface {
	retrieve(string) ([]byte, error)
}
type RetrieverFunc func(key string) ([]byte, error)

func (f RetrieverFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

type SaberCache struct {
	cache     *Cache
	server    *Server
	retriever Retriever
	flight    *singleflight.Flight
}

func NewSaberCache(maxBytes int64, strategy string, retriever Retriever) *SaberCache {
	if retriever == nil {
		panic("Group retriever must be existed!")
	}
	sc := &SaberCache{
		cache:     newCache(maxBytes, strategy),
		retriever: retriever,
		flight:    &singleflight.Flight{},
	}
	sabercache = sc
	sc.cache.Init()
	go sc.cache.BgSave()
	return sc
}
func (sc *SaberCache) RegisterSvr(svr *Server) {
	if sc.server != nil {
		panic("SaberCache had been registered server")
	}
	sc.server = svr
}
func (sc *SaberCache) Set(key string, value ByteView, ttl int64) bool {
	if ttl == -1 {
		sc.cache.SetWithoutTTL(key, value)
	} else {
		sc.cache.SetWithTTL(key, value, ttl)
	}
	return true
}
func (sc *SaberCache) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	if value, ok := sc.cache.Get(key); ok {
		log.Println("cache hit")
		return value, nil
	}
	return sc.load(key)
}
func (sc *SaberCache) GetAll() (kv []*cachememory.Entity) {
	return sc.cache.GetAll()
}
func (sc *SaberCache) Save() bool {
	return sc.cache.Save()
}

func (sc *SaberCache) TTL(key string) int64 {
	return sc.cache.TTL(key)
}
func (sc *SaberCache) load(key string) (ByteView, error) {
	view, err := sc.flight.Fly(key, func() (any, error) {
		return sc.getLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return view.(ByteView), nil

}

// getLocally 本地向Retriever取回数据并填充缓存
func (sc *SaberCache) getLocally(key string) (ByteView, error) {
	bytes, err := sc.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{bytes: cloneBytes(bytes)}
	sc.populateCache(key, value)
	return value, nil
}

// populateCache 提供填充缓存的能力
func (sc *SaberCache) populateCache(key string, value ByteView) {
	ttl := rand.Int63n(60)
	sc.cache.SetWithTTL(key, value, ttl)
}
