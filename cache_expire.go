package cache

import (
	"sync"
	"time"
)

/** I loved this idea of bucket the item according to expiration. I got it while reading the document of Ristretto,
a very good library for local cache */

type itemExpireData map[uint64]byte

type expirationData[T any] struct {
	sync.Mutex
	expirationBucket map[int64]itemExpireData
}

func expirationBucketKey(t time.Time) int64 {
	return (t.Unix() / ExpirationInterval) + 1
}

func newExpirationData[T any]() *expirationData[T] {
	return &expirationData[T]{
		expirationBucket: make(map[int64]itemExpireData),
	}
}

func (e *expirationData[T]) add(key uint64, expiration time.Time) {
	if expiration.IsZero() {
		return
	}

	if e == nil {
		return
	}

	bucketNum := expirationBucketKey(expiration)
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	bucket, ok := e.expirationBucket[bucketNum]
	if !ok {
		bucket = make(itemExpireData)
		e.expirationBucket[bucketNum] = bucket
	}
	bucket[key] = byte(1)
}

func (e *expirationData[T]) removeExpiredItem(c cacheOp[T]) {
	removeBucketKey := expirationBucketKey(time.Now()) - 1
	e.Mutex.Lock()
	defer e.Mutex.Unlock()
	bucket, ok := e.expirationBucket[removeBucketKey]
	if ok {
		for key, _ := range bucket {
			c.Del(key)
		}
	}
}
