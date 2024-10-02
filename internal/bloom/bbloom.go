// The MIT License (MIT)
// Copyright (c) 2014 Andreas Briese, eduToolbox@Bri-C GmbH, Sarstedt

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package bbloom

import (
	"log"
	"math"
	"unsafe"
)

// helper
var mask = []uint8{1, 2, 4, 8, 16, 32, 64, 128}

func getSize(ui64 uint64) (size uint64, exponent uint64) {
	if ui64 < uint64(512) {
		ui64 = uint64(512)
	}
	size = uint64(1)
	for size < ui64 {
		size <<= 1
		exponent++
	}
	return size, exponent
}

func calcSizeByWrongPositives(numEntries, wrongs float64) (uint64, uint64) {
	size := -1 * numEntries * math.Log(wrongs) / math.Pow(float64(0.69314718056), 2)
	locs := math.Ceil(float64(0.69314718056) * size / numEntries)
	return uint64(size), uint64(locs)
}

// NewBloomFilter returns a new bloomfilter.
func NewBloomFilter(params ...float64) (bloomfilter *Bloom) {
	var entries, locs uint64
	if len(params) == 2 {
		if params[1] < 1 {
			entries, locs = calcSizeByWrongPositives(params[0], params[1])
		} else {
			entries, locs = uint64(params[0]), uint64(params[1])
		}
	} else {
		log.Fatal("usage: New(float64(number_of_entries), float64(number_of_hashlocations))" +
			" i.e. New(float64(1000), float64(3)) or New(float64(number_of_entries)," +
			" float64(number_of_hashlocations)) i.e. New(float64(1000), float64(0.03))")
	}
	size, exponent := getSize(entries)
	bloomfilter = &Bloom{
		sizeExp: exponent,
		size:    size - 1,
		setLocs: locs,
		shift:   64 - exponent,
	}
	bloomfilter.Size(size)
	return bloomfilter
}

// Bloom filter
type Bloom struct {
	bitset  []uint64
	ElemNum uint64
	sizeExp uint64
	size    uint64
	setLocs uint64
	shift   uint64
}

// Add adds hash of a key to the bloomfilter.
func (bl *Bloom) Add(hash uint64) {
	h := hash >> bl.shift
	l := hash << bl.shift >> bl.shift
	for i := uint64(0); i < bl.setLocs; i++ {
		bl.Set((h + i*l) & bl.size)
		bl.ElemNum++
	}
}

// Has checks if bit(s) for entry hash is/are set,
// returns true if the hash was added to the Bloom Filter.
func (bl Bloom) Has(hash uint64) bool {
	h := hash >> bl.shift
	l := hash << bl.shift >> bl.shift
	for i := uint64(0); i < bl.setLocs; i++ {
		if !bl.IsSet((h + i*l) & bl.size) {
			return false
		}
	}
	return true
}

// Size makes Bloom filter with as bitset of size sz.
func (bl *Bloom) Size(sz uint64) {
	bl.bitset = make([]uint64, sz>>6)
}

// Clear resets the Bloom filter.
func (bl *Bloom) Clear() {
	for i := range bl.bitset {
		bl.bitset[i] = 0
	}
}

// Set sets the bit[idx] of bitset.
func (bl *Bloom) Set(idx uint64) {
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(&bl.bitset[idx>>6])) + uintptr((idx%64)>>3))
	*(*uint8)(ptr) |= mask[idx%8]
}

// IsSet checks if bit[idx] of bitset is set, returns true/false.
func (bl *Bloom) IsSet(idx uint64) bool {
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(&bl.bitset[idx>>6])) + uintptr((idx%64)>>3))
	r := ((*(*uint8)(ptr)) >> (idx % 8)) & 1
	return r == 1
}
