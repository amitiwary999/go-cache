package cache

import (
	"math"
	"math/rand"
	"time"
)

const seedLen = 3

type countmin struct {
	filterItemSize uint64
	filter         [][]uint64
	seed           []uint64
}

func minFunc(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func newCountMin(size uint64) *countmin {
	cm := &countmin{
		filterItemSize: size,
	}
	for i := 0; i < seedLen; i++ {
		cm.filter[i] = make([]uint64, size)
		cm.seed[i] = rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()
	}
	return cm
}

func (c *countmin) setKeyCount(hash uint64) {
	for i, s := range c.seed {
		index := (hash ^ s) % uint64(c.filterItemSize)
		c.filter[i][index] = c.filter[i][index] + 1
	}
}

func (c *countmin) getKeyCount(hash uint64) uint64 {
	var min uint64 = math.MaxUint64
	for i, s := range c.seed {
		index := (hash ^ s) % uint64(c.filterItemSize)
		min = minFunc(min, c.filter[i][index])
	}
	return min
}
