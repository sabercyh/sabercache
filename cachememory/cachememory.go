package cachememory

type CacheMemory interface {
	Get(key string) (Value, bool)
	SetWithoutTTL(key string, value Value)
	SetWithTTL(key string, value Value, expireSecond int64)
	ExpireKeyMonitor()
	MultiDeleteKey(keys []string, t int64)
	RemoveExpiredKey(key string)
	Remove()
	TTL(key string) int64
	Len() int
	Stop()
}
type Entity struct {
	key         string
	value       Value
	expiredTime int64
}
type Value interface {
	Len() int
}

const DelChCap int = 100

type DelCH struct {
	keys []string
	t    int64
}
type OnEliminated func(key string, value Value)
