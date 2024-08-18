package cache

import (
	"time"

	"github.com/cespare/xxhash/v2"
)

var (
	ExpirationInterval = int64(5)
)

type CacheConfig struct {
	Capacity   uint64
	CountBatch uint64
	FreqCount  uint64
}

type Cache[T any] struct {
	data          cacheOp[T]
	done          chan int
	cleanupTicker *time.Ticker
}

func NewCacheWithCapacity[T any](cConfig *CacheConfig, done chan int) *Cache[T] {
	timer := time.NewTicker(time.Duration(ExpirationInterval) * time.Second)
	cache := &Cache[T]{
		data:          NewCacheData[T](cConfig, done),
		done:          done,
		cleanupTicker: timer,
	}
	go cache.cleanUp()
	return cache
}

func (c *Cache[T]) cleanUp() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.data.RemoveExpiredItem()
		case <-c.done:
			return
		}
	}
}

func (c *Cache[T]) Close() {
	close(c.done)
}

func (c *Cache[T]) Reset() {
	c.data.Reset()
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) error {
	expiration := time.Now().Add(ttl)
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	c.data.Set(keyInt, value, expiration)
	return nil
}

func (c *Cache[T]) Get(key string) (T, error) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	return c.data.Get(keyInt)
}

func (c *Cache[T]) Delete(key string) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	c.data.Del(keyInt)
}
