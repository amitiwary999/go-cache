package cache

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

var keyPref string = "key"
var valuePref string = "value"

func timeGC() time.Duration {
	start := time.Now()
	runtime.GC()
	return time.Since(start)
}

func saveData(bgr *bigCacheRing, start, end int) {
	for i := start; i < end; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		value := fmt.Sprintf("%v-%v", valuePref, i)
		bgr.Save(key, value)
	}
}

func Get(bgr *bigCacheRing, start, end int) {
	for i := start; i < end; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		bgr.Get(key)
	}
}

func Delete(bgr *bigCacheRing, start, end int) {
	for i := start; i < end; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		bgr.Delete(key)
	}
}

func BenchmarkBigCache(b *testing.B) {
	bgc, err := NewBigCacheRing(100000, 1, 120*time.Hour)
	if err == nil {
		saveData(bgc, 0, 500000)
		Get(bgc, 0, 12000)
		Delete(bgc, 0, 2000)
		Delete(bgc, 13000, 15000)
		Get(bgc, 12000, 25000)
		saveData(bgc, 500000, 800000)
		runtime.GC()
		fmt.Printf("With a map of strings, GC took: %s\\n", timeGC())
		Get(bgc, 50000, 75000)
		Delete(bgc, 100000, 120000)
		Get(bgc, 110000, 130000)
	}
}
