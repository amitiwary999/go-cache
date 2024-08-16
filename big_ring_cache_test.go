package cache

import (
	"testing"
	"time"
)

func TestBigRingCache(t *testing.T) {
	ti := &TickerInfo{
		Interval: 15 * time.Minute,
		Hour:     16,
		Min:      14,
		Sec:      50,
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
		value1, err := bch.Get("key1")
		if err != nil {
			t.Fatalf("error fetch key1 %v \n", err)
		} else if value1 != "value1" {
			t.Fatalf("fetch wrong value for key key1")
		}
	}

	bch.Save("key1", "value1")
	bch.Save("key2", "value2")
	bch.Save("key3", "value3")
	bch.Save("key4", "value4")
	bch.Save("key5", "value5")
	bch.Save("key6", "value6")
	bch.Save("key7", "value7")

	value1, err := bch.Get("key1")
	if err != nil {
		t.Fatalf("error fetch key1 %v \n", err)
	} else if value1 != "value1" {
		t.Fatalf("fetch wrong value for key key1")
	}

	value2, err := bch.Get("key2")
	if err != nil {
		t.Fatalf("error fetch key2 %v \n", err)
	} else if value2 != "value2" {
		t.Fatalf("fetch wrong value for key key2")
	}
	bch.Get("key3")
	bch.Get("key4")
	bch.Get("key5")
	value2, err = bch.Get("key2")
	if err != nil {
		t.Fatalf("error fetch key2 %v \n", err)
	} else if value2 != "value2" {
		t.Fatalf("fetch wrong value for key key2")
	}
	bch.Delete("key3")
	_, err3 := bch.Get("key3")
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}
	time.Sleep(5 * time.Second)
	bch.Clear()
	bch, initErr = NewBigCacheRing(5, ti)
	if initErr != nil {
		t.Errorf("failed to init cache %v \n", initErr)
	}
	bch.LoadFileOffset(doneCh)
	<-doneCh
	_, err3 = bch.Get("key3")
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}

}
