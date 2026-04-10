package cache

import (
	"sync"
	"time"
)

type item[T any] struct {
	value      T
	expiration int64
}

type TTLMap[T any] struct {
	sync.RWMutex
	items map[string]item[T]
}

func NewTTLMap[T any](cleaningInterval time.Duration) *TTLMap[T] {
	tm := &TTLMap[T]{
		items: make(map[string]item[T]),
	}
	if cleaningInterval > 0 {
		go tm.startCleaning(cleaningInterval)
	}
	return tm
}

func (tm *TTLMap[T]) Set(key string, value T, ttl time.Duration) {
	tm.Lock()
	defer tm.Unlock()
	tm.items[key] = item[T]{
		value:      value,
		expiration: time.Now().Add(ttl).UnixNano(),
	}
}

func (tm *TTLMap[T]) Get(key string) (T, bool) {
	tm.RLock()
	defer tm.RUnlock()
	it, ok := tm.items[key]
	if !ok {
		var zero T
		return zero, false
	}
	if time.Now().UnixNano() > it.expiration {
		var zero T
		return zero, false
	}
	return it.value, true
}

func (tm *TTLMap[T]) Delete(key string) {
	tm.Lock()
	defer tm.Unlock()
	delete(tm.items, key)
}

func (tm *TTLMap[T]) Clear() {
	tm.Lock()
	defer tm.Unlock()
	tm.items = make(map[string]item[T])
}

func (tm *TTLMap[T]) startCleaning(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		tm.Lock()
		now := time.Now().UnixNano()
		for k, v := range tm.items {
			if now > v.expiration {
				delete(tm.items, k)
			}
		}
		tm.Unlock()
	}
}
