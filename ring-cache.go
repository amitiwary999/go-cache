package cache

import (
	"container/ring"
)

type cacheRing struct {
	data map[string]*ring.Ring
	hand *ring.Ring
}

type cacheItem[T any] struct {
	item T
}

func NewCacheRing(capacity int32) *cacheRing {
	r := ring.New(int(capacity))
	return &cacheRing{
		data: make(map[string]*ring.Ring),
		hand: r,
	}
}
