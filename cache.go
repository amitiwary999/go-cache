package cache

import (
	"time"

	"github.com/cespare/xxhash/v2"
)

var (
	stopTicker         chan int
	ExpirationInterval = int64(5)
)

type Cache[T any] struct {
	data          cacheOp[T]
	done          chan int
	capacity      int16
	cleanupTicker *time.Ticker
}

func NewCacheWithCapacity[T any](capacity int16, countBatch uint64, done chan int) *Cache[T] {
	timer := time.NewTicker(time.Duration(ExpirationInterval) * time.Second)
	cache := &Cache[T]{
		data:          NewCacheData[T](countBatch, done),
		capacity:      capacity,
		done:          done,
		cleanupTicker: timer,
	}
	go cache.cleanUp()
	return cache
}

func (c *Cache[T]) StopCleanUp() {
	c.cleanupTicker.Stop()
	stopTicker <- 1
}

func (c *Cache[T]) cleanUp() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.data.RemoveExpiredItem()
		case <-c.done:
			return
		case <-stopTicker:
			return
		}
	}
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
