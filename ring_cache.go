package cache

import (
	"container/ring"
	"errors"

	xxhash "github.com/cespare/xxhash/v2"
)

type cacheRing[T any] struct {
	data map[uint64]*ring.Ring
	hand *ring.Ring
}

type cacheRingItem[T any] struct {
	key       uint64
	item      T
	reference int8
}

func NewCacheRing[T any](capacity int32) *cacheRing[T] {
	r := ring.New(int(capacity))
	return &cacheRing[T]{
		data: make(map[uint64]*ring.Ring),
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

func (c *cacheRing[T]) Set(key string, value T) error {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	ringVal, ok := c.data[keyInt]
	if ok {
		cacheItem := ringVal.Value.(*cacheRingItem[T])
		cacheItem.key = keyInt
		cacheItem.item = value
		cacheItem.reference = 0
		return nil
	}
	item := &cacheRingItem[T]{
		key:       keyInt,
		item:      value,
		reference: 0,
	}
	if c.hand.Value == nil {
		c.hand.Value = item
		c.data[keyInt] = c.hand
	} else {
		c.findReplaceItem()
		c.hand.Value = item
		c.data[keyInt] = c.hand
	}
	c.hand = c.hand.Next()
	return nil
}

func (c *cacheRing[T]) Get(key string) (T, error) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	var itemVal T
	ringVal, ok := c.data[keyInt]
	if ok {
		cacheItem := ringVal.Value.(*cacheRingItem[T])
		cacheItem.reference = 1
		return cacheItem.item, nil
	} else {
		return itemVal, errors.New("key not found")
	}
}

func (c *cacheRing[T]) Delete(key string) error {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	ringVal, ok := c.data[keyInt]
	if ok {
		prevHand := ringVal.Prev()
		nextHand := ringVal.Next()
		prevHand.Link(nextHand)
		ringVal.Value = nil
		delete(c.data, keyInt)
		return nil
	} else {
		return errors.New("key not found")
	}
}
