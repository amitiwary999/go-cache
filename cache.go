package inmemeorycache

import (
	"errors"
	"time"
)

type CacheItem[T any] struct {
	item       T
	expiration int64
}

type Cache[T any] struct {
	data            map[string]CacheItem[T]
	cleanupInterval int16
	done            chan int
	capacity        int16
}

func NewCacheWithCapacity[T any](capacity, cleanupInterval int16, done chan int) *Cache[T] {
	cache := &Cache[T]{
		data:            make(map[string]CacheItem[T], capacity),
		cleanupInterval: cleanupInterval,
		capacity:        capacity,
		done:            done,
	}
	cache.cleanUp()
	return cache
}

func NewCache[T any](cleanupInterval int16, done chan int) *Cache[T] {
	cache := &Cache[T]{
		data:            make(map[string]CacheItem[T]),
		cleanupInterval: cleanupInterval,
		done:            done,
	}
	cache.cleanUp()
	return cache
}

func (c *Cache[T]) deleteExpiredItem() {
	currentSecond := time.Now().Unix()
	for k, v := range c.data {
		if v.expiration >= currentSecond {
			delete(c.data, k)
		}
	}
}

func (c *Cache[T]) cleanUp() {
	timer := time.NewTicker(time.Duration(c.cleanupInterval) * time.Second)
	for {
		select {
		case <-timer.C:
			c.deleteExpiredItem()
		case <-c.done:
			return
		}
	}
}

func (c *Cache[T]) Set(key string, value T, expirationSecond int64) error {
	item := CacheItem[T]{
		item:       value,
		expiration: expirationSecond,
	}
	if int(c.capacity) > 0 && len(c.data) >= int(c.capacity) {
		return errors.New("cache is full")
	}
	c.data[key] = item
	return nil
}

func (c *Cache[T]) Get(key string) (T, error) {
	var value T
	item, ok := c.data[key]
	if ok {
		value = item.item
		return value, nil
	} else {
		return value, errors.New("key not found in cache")
	}
}
