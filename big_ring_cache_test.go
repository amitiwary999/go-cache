package cache

import (
	"fmt"
	"testing"
)

func TestBigRingCache(t *testing.T) {
	bch, initErr := NewBigCacheRing(5)
	if initErr != nil {
		fmt.Printf("failed to init cache %v \n", initErr)
	} else {
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
	}
}
