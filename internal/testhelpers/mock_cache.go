package testhelpers

import (
	"errors"
	"time"

	"github.com/rezect/url-shortener/internal/cache"
)

type MockCache struct{}

func (c *MockCache) Get(key string) (string, error) {
	switch key {
	case "existsCache":
		return "original link", cache.CacheHit
	case "notExistsCache":
		return "", cache.NotExists
	default:
		return "", cache.CacheMiss
	}
}

func (c *MockCache) Set(key string, value any, ttl time.Duration) error {
	return nil
}

func (c *MockCache) Delete(key string) error {
	if key == "exists-cache" {
		return nil
	} else {
		return errors.New("key not found")
	}
}

func (c *MockCache) Clear() {}

func (c *MockCache) Stop() {}
