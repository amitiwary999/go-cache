package cache

import (
	"sync"
	"time"
)

/** I loved this idea of bucket the item according to expiration. I got it while reading the document of Ristretto,
a very good library for local cache */

type itemExpireData []uint64

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
	bucket, ok := e.expirationBucket[bucketNum]
	if !ok {
		bucket = make(itemExpireData, 100)
		e.expirationBucket[bucketNum] = bucket
	}
	bucket = append(bucket, key)
}

func (e *expirationData[T]) removeExpiredItem(c CacheData[T]) {
	removeBucketKey := expirationBucketKey(time.Now()) - 1
	bucket, ok := e.expirationBucket[removeBucketKey]
	if ok {
		for _, key := range bucket {
			c.Del(key)
		}
	}
}
