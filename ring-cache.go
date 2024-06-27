package cache

import (
	"container/ring"
	"errors"
)

type cacheRing[T any] struct {
	data map[string]*ring.Ring
	hand *ring.Ring
}

type cacheRingItem[T any] struct {
	key       string
	item      T
	reference int8
}

func NewCacheRing[T any](capacity int32) *cacheRing[T] {
	r := ring.New(int(capacity))
	return &cacheRing[T]{
		data: make(map[string]*ring.Ring),
		hand: r,
	}
}

func (c *cacheRing[T]) findReplaceItem() {
	for c.hand.Value.(*cacheRingItem[T]).reference == 1 {
		c.hand.Value.(*cacheRingItem[T]).reference = 0
		c.hand = c.hand.Next()
	}
	if c.hand.Value != nil {
		currentItem := c.hand.Value.(*cacheRingItem[T])
		delete(c.data, currentItem.key)
		c.hand.Value = nil
	}
}

func (c *cacheRing[T]) Set(key string, value T) {
	item := &cacheRingItem[T]{
		key:       key,
		item:      value,
		reference: 0,
	}
	if c.hand.Value == nil {
		c.hand.Value = item
		c.data[key] = c.hand
	} else {
		c.findReplaceItem()
		c.hand.Value = item
		c.data[key] = c.hand
	}
	c.hand = c.hand.Next()
}

func (c *cacheRing[T]) Get(key string) (T, error) {
	var itemVal T
	ringVal, ok := c.data[key]
	if ok {
		cacheItem := ringVal.Value.(*cacheRingItem[T])
		cacheItem.reference = 1
		return cacheItem.item, nil
	} else {
		return itemVal, errors.New("key not found")
	}
}
