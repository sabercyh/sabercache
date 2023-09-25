package singleflight

import (
	"sync"
)

type packet struct {
	wg  sync.WaitGroup
	val any
	err error
}
type Flight struct {
	mu     sync.Mutex
	flight map[string]*packet
}

func (f *Flight) Fly(key string, fn func() (any, error)) (any, error) {
	f.mu.Lock()
	if f.flight == nil {
		f.flight = make(map[string]*packet)
	}
	if p, ok := f.flight[key]; ok {
		f.mu.Unlock()
		p.wg.Wait()
		return p.val, p.err
	}
	p := new(packet)
	p.wg.Add(1)
	f.flight[key] = p
	f.mu.Unlock()

	p.val, p.err = fn()
	p.wg.Done()

	f.mu.Lock()
	delete(f.flight, key)
	f.mu.Unlock()

	return p.val, p.err

}
