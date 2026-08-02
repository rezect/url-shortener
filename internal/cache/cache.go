package cache

import (
	"errors"
	"time"

	"github.com/rezect/ttl-cache-server/pkg/cache"
)

var (
	CacheMiss = errors.New("cache miss")
	CacheHit = errors.New("cache hit")
	NotExists = errors.New("not exists")
	ErrWrongValue = errors.New("not a string contains in value")
)

type Cache struct {
	cache *cache.Cache
}

func NewCache() *Cache {
	return &Cache{
		cache: cache.CacheNew(),
	}
}

func (c *Cache) Get(key string) (string, error) {
	value, err := c.cache.Get(key)
	if err != nil {
		return "", CacheMiss
	}
	if value == nil {
		return "", NotExists
	}
	if strValue, ok := value.(string); ok {
		return strValue, CacheHit		
	} else {
		return "", ErrWrongValue
	}
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	return c.cache.Set(key, value, ttl)
}

func (c *Cache) Delete(key string) error {
	return c.cache.Delete(key)
}

func (c *Cache) Clear() {
	c.cache.Clear()
}

func (c *Cache) Stop() {
	c.cache.Stop()
}
