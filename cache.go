package inmemeorycache

import "time"

type CacheItem[T any] struct {
	item       T
	expiration int64
}

type InMemoryCache[T any] struct {
	data            map[string]CacheItem[T]
	cleanupInterval uint16
	done            chan int
}

func NewInMemoryCache[T any](initialCapacity, cleanupInterval uint16, done chan int) *InMemoryCache[T] {
	cache := &InMemoryCache[T]{
		data:            make(map[string]CacheItem[T], initialCapacity),
		cleanupInterval: cleanupInterval,
		done:            done,
	}
	cache.cleanUp()
	return cache
}

func (c *InMemoryCache[T]) deleteExpiredItem() {
	currentSecond := time.Now().Unix()
	for k, v := range c.data {
		if v.expiration >= currentSecond {
			delete(c.data, k)
		}
	}
}

func (c *InMemoryCache[T]) cleanUp() {
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

func (c *InMemoryCache[T]) Set(key string, value T, expirationSecond int64) {
	item := CacheItem[T]{
		item:       value,
		expiration: expirationSecond,
	}
	c.data[key] = item
}
