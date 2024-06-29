package cache

import "os"

type bigCacheRing[T any] struct {
	file      *os.File
	offsetMap map[string]int64
	cacheRing *cacheRing[T]
}

func NewBigCacheRing[T any]() (*bigCacheRing[T], error) {
	file, err := os.OpenFile("big-cache-ring-data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	cacheRing := NewCacheRing[T](1000)
	offsetMap := make(map[string]int64)
	return &bigCacheRing[T]{
		file:      file,
		offsetMap: offsetMap,
		cacheRing: cacheRing,
	}, nil
}
