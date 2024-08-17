package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestBigRingCache(t *testing.T) {
	ti := &TickerInfo{
		Interval: 5 * time.Second,
		Hour:     23,
		Min:      17,
		Sec:      5,
	}
	bch, initErr := NewBigCacheRing(5, ti)
	if initErr != nil {
		t.Errorf("failed to init cache %v \n", initErr)
	}
	doneCh := make(chan int)
	bch.LoadFileOffset(doneCh)
	<-doneCh
	fileCacheSize := bch.Size(1)
	t.Logf("already present file cache size %v \n", fileCacheSize)
	if fileCacheSize > 0 {
		/** We don't delete the key1 in this test so we assume that on file load the key1 should be present*/
		value1, err := bch.Get(fmt.Sprintf("%v-%v", keyPref, 1))
		if err != nil {
			t.Fatalf("error fetch key1 %v \n", err)
		} else if value1 != fmt.Sprintf("%v-%v", valuePref, 1) {
			t.Fatalf("fetch wrong value for key key1")
		}
	}

	saveData(bch, 1, 100000)

	value1, err := bch.Get(fmt.Sprintf("%v-%v", keyPref, 1))
	if err != nil {
		t.Fatalf("error fetch key1 %v \n", err)
	} else if value1 != fmt.Sprintf("%v-%v", valuePref, 1) {
		t.Fatalf("fetch wrong value for key key1")
	}

	value2, err := bch.Get(fmt.Sprintf("%v-%v", keyPref, 2))
	if err != nil {
		t.Fatalf("error fetch key2 %v \n", err)
	} else if value2 != fmt.Sprintf("%v-%v", valuePref, 2) {
		t.Fatalf("fetch wrong value for key key2")
	}
	Get(bch, 3, 200)
	value2, err = bch.Get(fmt.Sprintf("%v-%v", keyPref, 2))
	if err != nil {
		t.Fatalf("error fetch key2 %v \n", err)
	} else if value2 != fmt.Sprintf("%v-%v", valuePref, 2) {
		t.Fatalf("fetch wrong value for key key2")
	}
	bch.Delete(fmt.Sprintf("%v-%v", keyPref, 3))
	_, err3 := bch.Get(fmt.Sprintf("%v-%v", keyPref, 3))
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}
	saveData(bch, 200000, 200010)
	time.Sleep(5 * time.Second)
	saveData(bch, 200010, 900000)
	bch.Clear()
	bch, initErr = NewBigCacheRing(5, ti)
	if initErr != nil {
		t.Errorf("failed to init cache %v \n", initErr)
	}
	bch.LoadFileOffset(doneCh)
	<-doneCh
	_, err3 = bch.Get(fmt.Sprintf("%v-%v", keyPref, 3))
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}

}
