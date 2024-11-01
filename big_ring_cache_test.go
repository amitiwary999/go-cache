package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestBigRingCacheDelete(t *testing.T) {
	currentTime := time.Now()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	second := currentTime.Second()
	second = second + 3
	if (second / 60) >= 1 {
		second = second % 60
		minute = minute + 1
		if (minute / 60) >= 1 {
			minute = minute % 60
			hour = hour + 1
			if (hour / 24) >= 1 {
				hour = hour % 24
			}
		}
	}

	ti := &TickerInfo{
		Interval: 55 * time.Second,
		Hour:     hour,
		Min:      minute,
		Sec:      second,
	}
	bch, initErr := NewBigCacheRing(5, ti)
	if initErr != nil {
		t.Errorf("failed to init cache %v \n", initErr)
	}
	doneCh := make(chan int)

	saveData(bch, 1, 500000)

	bch.Delete(fmt.Sprintf("%v-%v", keyPref, 3))
	_, err3 := bch.Get(fmt.Sprintf("%v-%v", keyPref, 3))
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}
	time.Sleep(2999 * time.Millisecond)
	saveData(bch, 500010, 2000000)
	/** clean data and init. This to test that the deleted keys are removed from file. */
	bch.Clear()
	bch, initErr = NewBigCacheRing(5, ti)
	if initErr != nil {
		t.Errorf("failed to init cache %v \n", initErr)
	}
	bch.LoadFileOffset(doneCh)
	bch.Delete(fmt.Sprintf("%v-%v", keyPref, 4))
	<-doneCh
	_, err3 = bch.Get(fmt.Sprintf("%v-%v", keyPref, 3))
	if err3 == nil || err3.Error() != "key not found" {
		t.Fatalf("key3 is deleted so no value should be present")
	}

}

func TestBigRingCache(t *testing.T) {
	currentTime := time.Now()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	second := currentTime.Second()
	second = second + 3
	if (second / 60) >= 1 {
		second = second % 60
		minute = minute + 1
		if (minute / 60) >= 1 {
			minute = minute % 60
			hour = hour + 1
			if (hour / 24) >= 1 {
				hour = hour % 24
			}
		}
	}

	ti := &TickerInfo{
		Interval: 15 * time.Second,
		Hour:     hour,
		Min:      minute,
		Sec:      second,
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

	saveData(bch, 1, 500)

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
}
