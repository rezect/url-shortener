package testhelpers

import (
	"errors"
	"time"
)

type MockCache struct{}

func (c *MockCache) Get(key string) (any, error) {
	if key == "exists" {
		return "some link", nil
	} else {
		return "", errors.New("key not found")
	}
}

func (c *MockCache) Set(key string, value any, ttl time.Duration) error {
	return nil
}

func (c *MockCache) Delete(key string) error {
	if key == "exists" {
		return nil
	} else {
		return errors.New("key not found")
	}
}

func (c *MockCache) Clear() {}

func (c *MockCache) Stop() {}