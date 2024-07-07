package cache

import (
	"fmt"
	"testing"
)

var keyPref string = "key"
var valuePref string = "value"

func saveData(bgr *bigCacheRing) {
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		value := fmt.Sprintf("%v-%v", valuePref, i)
		bgr.Save(key, value)
	}
}

func Get(bgr *bigCacheRing) {
	for i := 0; i < 25000; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		bgr.Get(key)
	}
}

func Delete(bgr *bigCacheRing) {
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("%v-%v", keyPref, i)
		bgr.Delete(key)
	}
}

func BenchmarkBigCache(b *testing.B) {
	bgc, err := NewBigCacheRing(10000)
	if err == nil {
		for i := 0; i < 100000; i++ {
			key := fmt.Sprintf("%v-%v", keyPref, i)
			value := fmt.Sprintf("%v-%v", valuePref, i)
			bgc.Save(key, value)
		}

		for i := 0; i < 25000; i++ {
			key := fmt.Sprintf("%v-%v", keyPref, i)
			bgc.Get(key)
		}

		saveData(bgc)
		Delete(bgc)
		Get(bgc)
	}
}
