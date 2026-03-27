package infra

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache provides in-memory caching with TTL support backed by go-cache.
type Cache struct {
	c *gocache.Cache
}

// NewCache creates a new cache with 5-minute default expiry and 10-minute cleanup interval.
func NewCache() *Cache {
	return &Cache{c: gocache.New(5*time.Minute, 10*time.Minute)}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.c.Get(key)
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.c.Set(key, value, ttl)
}

func (c *Cache) Delete(key string) {
	c.c.Delete(key)
}

func (c *Cache) Clear() {
	c.c.Flush()
}

func (c *Cache) Size() int {
	return c.c.ItemCount()
}
