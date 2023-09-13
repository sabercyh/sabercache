package cachememory

type CacheMemory interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	Remove()
	Len() int
}
type Entity struct {
	key   string
	value Value
}
type Value interface {
	Len() int
}
type OnEliminated func(key string, value Value)
