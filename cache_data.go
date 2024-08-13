package cache

import (
	"errors"
	"sync"
	"time"

	cacheheap "github.com/amitiwary999/go-cache/internal/heap"
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
	sync.Mutex
	data           []*cacheDataMap[T]
	getCountBatch  []uint64
	capacity       uint64
	size           int64
	itemsCh        chan []uint64
	done           chan int
	batchSize      uint64
	lfuSketch      countmin
	lfuQueue       PriorityQueue
	queuePosMap    map[uint64]int
	expirationData *expirationData[T]
}

type cacheOp[T any] interface {
	Set(uint64, T, time.Time)
	Get(uint64) (T, error)
	Del(uint64)
	Reset()
	RemoveExpiredItem()
}

func newCacheDataMap[T any]() *cacheDataMap[T] {
	c := &cacheDataMap[T]{
		dataMap: make(map[uint64]cacheItem[T]),
	}
	return c
}

func NewCacheData[T any](cConfig *CacheConfig, done chan int) cacheOp[T] {
	c := &CacheData[T]{
		data:           make([]*cacheDataMap[T], 256),
		getCountBatch:  make([]uint64, cConfig.CountBatch),
		capacity:       cConfig.Capacity,
		itemsCh:        make(chan []uint64, 5),
		batchSize:      cConfig.CountBatch,
		done:           done,
		lfuSketch:      *newCountMin(cConfig.FreqCount),
		lfuQueue:       make(PriorityQueue, 0),
		queuePosMap:    make(map[uint64]int),
		expirationData: newExpirationData[T](),
	}
	for i := range c.data {
		c.data[i] = newCacheDataMap[T]()
	}
	cacheheap.Init(&c.lfuQueue)
	go c.process()
	return c
}

func (c *CacheData[T]) close() {
	close(c.itemsCh)
}

func (c *CacheData[T]) Set(key uint64, value T, expiration time.Time) {
	i := key % 256
	item := cacheItem[T]{
		item:       value,
		expiration: expiration,
	}
	oldItem, update := c.data[i].set(key, item)
	if update {
		c.expirationData.update(key, oldItem.expiration, expiration)
	} else {
		c.expirationData.add(key, expiration)
		c.addFreq(key)
		c.changeSize(1)
		c.removeExcessItem()
	}
}

func (c *cacheDataMap[T]) set(key uint64, item cacheItem[T]) (cacheItem[T], bool) {
	c.Lock()
	defer c.Unlock()
	oldItem, ok := c.dataMap[key]
	c.dataMap[key] = item
	return oldItem, ok
}

func (c *CacheData[T]) Get(key uint64) (T, error) {
	i := key % 256
	c.addFreq(key)
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
	ok := c.data[i].del(key)
	if ok {
		c.changeSize(-1)
	}
}

func (c *cacheDataMap[T]) del(key uint64) bool {
	_, ok := c.dataMap[key]
	delete(c.dataMap, key)
	return ok
}

func (c *CacheData[T]) RemoveExpiredItem() {
	c.expirationData.removeExpiredItem(c)
}

func (c *CacheData[T]) addFreq(key uint64) {
	c.getCountBatch = append(c.getCountBatch, key)
	if len(c.getCountBatch) >= int(c.batchSize) {
		c.itemsCh <- c.getCountBatch
		c.getCountBatch = c.getCountBatch[:0]
	}
}

func (c *CacheData[T]) removeExcessItem() {
	c.Lock()
	defer c.Unlock()
	if c.size >= int64(c.capacity) {
		lfuItemI := cacheheap.Pop(&c.lfuQueue)
		if lfuItemI != nil {
			lfuItem := lfuItemI.((*LFUItem))
			c.Del(lfuItem.key)
		}
	}
}

func (c *CacheData[T]) process() {
	for {
		select {
		case items := <-c.itemsCh:
			uniqueItems := make(map[uint64]struct{})
			for _, item := range items {
				c.lfuSketch.setKeyCount(item)
				uniqueItems[item] = struct{}{}
			}
			for k, _ := range uniqueItems {
				freq := c.lfuSketch.getKeyCount(k)
				item := &LFUItem{
					key:  k,
					freq: freq,
				}
				oldPos, pOk := c.queuePosMap[k]
				if pOk {
					c.lfuQueue.update(item, oldPos)
					cacheheap.Fix(&c.lfuQueue, oldPos)
				} else {
					pos := cacheheap.Push(&c.lfuQueue, item)
					c.queuePosMap[k] = pos
				}
			}
			uniqueItems = nil
		case <-c.done:
			c.close()
			return
		}
	}
}

func (c *CacheData[T]) changeSize(changeSize int64) {
	c.size = c.size + changeSize
}

func (c *CacheData[T]) Reset() {
	c.lfuQueue.reset()
	c.lfuSketch.reset()
}
