package cache

import "github.com/cespare/xxhash/v2"

type cacheItem[T any] struct {
	item       T
	expiration int64
}

type cacheDataMap[T any] struct {
	dataMap map[uint64]cacheItem[T]
}

type CacheData[T any] struct {
	data []*cacheDataMap[T]
}

func newCacheDataMap[T any]() *cacheDataMap[T] {
	c := &cacheDataMap[T]{
		dataMap: make(map[uint64]cacheItem[T]),
	}
	return c
}

func NewCacheData[T any]() *CacheData[T] {
	c := &CacheData[T]{
		data: make([]*cacheDataMap[T], 256),
	}
	for i := range c.data {
		c.data[i] = newCacheDataMap[T]()
	}
	return c
}

func (c *CacheData[T]) Set(key string, value T) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
}
