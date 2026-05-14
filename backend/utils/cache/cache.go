package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with an optional expiration time.
type Item[T any] struct {
	Value      T
	Expiration int64
}

// Cache is a simple thread-safe in-memory cache.
type Cache[T any] struct {
	items map[string]Item[T]
	mu    sync.RWMutex
}

// New creates a new Cache.
func New[T any]() *Cache[T] {
	return &Cache[T]{
		items: make(map[string]Item[T]),
	}
}

// Set adds an item to the cache. If duration is 0, the item never expires.
func (c *Cache[T]) Set(key string, value T, duration time.Duration) {
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item[T]{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		var zero T
		return zero, false
	}

	return item.Value, true
}

// Delete removes an item from the cache.
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush clears all items from the cache.
func (c *Cache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item[T])
}
