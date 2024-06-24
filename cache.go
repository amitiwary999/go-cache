package inmemeorycache

import (
	"errors"
	"time"
)

var stopTicker chan int

type CacheItem[T any] struct {
	item       T
	expiration int64
}

type Cache[T any] struct {
	data          map[string]CacheItem[T]
	done          chan int
	capacity      int16
	cleanupTicker *time.Ticker
}

func NewCacheWithCapacity[T any](capacity, cleanupInterval int16, done chan int) *Cache[T] {
	timer := time.NewTicker(time.Duration(cleanupInterval) * time.Second)
	cache := &Cache[T]{
		data:          make(map[string]CacheItem[T], capacity),
		capacity:      capacity,
		done:          done,
		cleanupTicker: timer,
	}
	cache.cleanUp()
	return cache
}

func NewCache[T any](cleanupInterval int16, done chan int) *Cache[T] {
	timer := time.NewTicker(time.Duration(cleanupInterval) * time.Second)
	cache := &Cache[T]{
		data:          make(map[string]CacheItem[T]),
		done:          done,
		cleanupTicker: timer,
	}
	cache.cleanUp()
	return cache
}

func (c *Cache[T]) StopCleanUp() {
	c.cleanupTicker.Stop()
	stopTicker <- 1
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
	for {
		select {
		case <-c.cleanupTicker.C:
			c.deleteExpiredItem()
		case <-c.done:
			return
		case <-stopTicker:
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

func (c *Cache[T]) Delete(key string) {
	delete(c.data, key)
}
