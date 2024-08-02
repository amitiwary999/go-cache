package cache

import (
	"errors"
	"sync"
	"time"
)

type cacheItem[T any] struct {
	item       T
	expiration time.Time
}

type cacheDataMap[T any] struct {
	sync.RWMutex
	dataMap map[uint64]cacheItem[T]
}

type CacheData[T any] struct {
	data           []*cacheDataMap[T]
	expirationData *expirationData[T]
}

func newCacheDataMap[T any]() *cacheDataMap[T] {
	c := &cacheDataMap[T]{
		dataMap: make(map[uint64]cacheItem[T]),
	}
	return c
}

func NewCacheData[T any]() *CacheData[T] {
	c := &CacheData[T]{
		data:           make([]*cacheDataMap[T], 256),
		expirationData: newExpirationData[T](),
	}
	for i := range c.data {
		c.data[i] = newCacheDataMap[T]()
	}
	return c
}

func (c *CacheData[T]) Set(key uint64, value T, expiration time.Time) {
	i := key % 256
	item := cacheItem[T]{
		item:       value,
		expiration: expiration,
	}
	c.data[i].set(key, item)
}

func (c *cacheDataMap[T]) set(key uint64, item cacheItem[T]) {
	c.Lock()
	defer c.Unlock()
	c.dataMap[key] = item
}

func (c *CacheData[T]) Get(key uint64) (T, error) {
	i := key % 256
	return c.data[i].get(key)
}

func (c *cacheDataMap[T]) get(key uint64) (T, error) {
	var item T
	c.RLock()
	defer c.RUnlock()
	cacheItem, ok := c.dataMap[key]
	if !ok {
		return item, errors.New("item not found")
	}
	if !cacheItem.expiration.IsZero() && time.Now().After(cacheItem.expiration) {
		return item, errors.New("item has expired")
	}
	return cacheItem.item, nil
}

func (c *CacheData[T]) Del(key uint64) {
	i := key % 256
	c.data[i].del(key)
}

func (c *cacheDataMap[T]) del(key uint64) {
	delete(c.dataMap, key)
}
