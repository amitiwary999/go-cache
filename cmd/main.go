package main

import (
	"fmt"
	"time"

	cache "github.com/amitiwary999/go-cache"
)

func main() {
	done := make(chan int)
	nch := cache.NewCacheWithCapacity[int](&cache.CacheConfig{
		Capacity:   20,
		CountBatch: 5,
		FreqCount:  10,
	}, done)
	nch.Set("hi", 5, time.Duration(5*time.Second))
	val, err := nch.Get("hi")
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Printf("first fetch %v \n", val)
	}
	nch.Set("1", 1, time.Duration(5*time.Second))
	nch.Set("2", 2, time.Duration(5*time.Second))
	nch.Set("3", 3, time.Duration(5*time.Second))
	nch.Set("4", 4, time.Duration(12*time.Second))
	nch.Set("5", 5, time.Duration(5*time.Second))
	nch.Get("2")
	nch.Get("3")
	nch.Get("3")
	nch.Get("4")
	nch.Get("4")
	nch.Get("3")
	nch.Set("1", 3, time.Duration(12*time.Second))
	vl1, _ := nch.Get("1")
	fmt.Printf("second fetch 1 %v \n", vl1)
	time.Sleep(time.Duration(1500 * time.Millisecond))
	for i := 6; i <= 12; i++ {
		nch.Set(fmt.Sprint(i), i, time.Duration(5*time.Second))
	}
	time.Sleep(time.Duration(300 * time.Millisecond))
	for i := 1; i <= 10; i++ {
		vl, _ := nch.Get(fmt.Sprint(i))
		fmt.Printf("value of the key %v is %v \n", i, vl)
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
	for i := 13; i <= 25; i++ {
		time.Sleep(time.Duration(50 * time.Millisecond))
		nch.Set(fmt.Sprint(i), i, time.Duration(5*time.Second))
	}
	time.Sleep(time.Duration(300 * time.Millisecond))
	for i := 1; i <= 25; i++ {
		vl, _ := nch.Get(fmt.Sprint(i))
		fmt.Printf("value of the key %v is %v \n", i, vl)
	}
}
