# go-cache
Save data on machine and fetch fast. Local cache.
Sometimes we need to save the data in the machine where we want to use that. This makes the fetching data fast. We may want that the data persist even when machine restart or it should not persist.This library has both option.

Library support two type of cache with the clock cache replacement .
1. All the data in the memory
2. Save all data in file in disk and load some data in memory for fast access

Get the library:<br> go get -u github.com/amitiwary999/go-cache<br>
Import library:<br> 
```
import (cache "github.com/amitiwary999/go-cache")
```
<br>

init first type of cache<br>
```
cacheRing := cach.NewCacheRing[type](sizeOfCache)
```
<br>
type is the data type of the key of cache. Like int16, int32, string etc.<br>
sizeOfCache is the number of keys of cache.<br>
<br>
init second type of cahce
<br>

```
bigCache := cache.NewBigCacheRing(sizeOfCache)
```
sizeOfCache is the number of keys of cache present in memeory(this is not the limit for the keys we save in file).<br>
<br>

### Save data 
```
cacheRing.Set(key,value)
or
bigCache.Set(key,value)
```
It returns error if failed to save data.

<br>

### Get Data

```
cacheRing.Get(key)
or
bigCache.Get(key)
```
It returns data and error. 
<br>

### To remove the key
```
cacheRing.Delete(key)
or
bigCache.Delete(key)
```
<br>
## Benchmark:

JLMP250:go-cache amitt$ go test -timeout 30m  -bench=.  -benchmem -memprofile memprofile.out -cpuprofile profile.out  -benchtime=20s -count=5
goos: darwin
goarch: arm64
pkg: github.com/amitiwary999/go-cache
BenchmarkBigCache-10    	1000000000	         2.102 ns/op	       0 B/op	       0 allocs/op
BenchmarkBigCache-10    	1000000000	         2.139 ns/op	       0 B/op	       0 allocs/op
BenchmarkBigCache-10    	1000000000	         2.172 ns/op	       0 B/op	       0 allocs/op

When I tried with map size 400000 there is almost no allocation. 
But when I increase the size and save the 800000 key on map the result was 

BenchmarkBigCache-10    	       1	2409240291 ns/op	206634832 B/op	 5508270 allocs/op
BenchmarkBigCache-10    	       1	2371712709 ns/op	206603024 B/op	 5508174 allocs/op
BenchmarkBigCache-10    	       1	2471930375 ns/op	206604184 B/op	 5508217 allocs/op

See the allocation. Suddenly there is lot of allocation and most happening in save where we use map to save the key offset.

*Tracking GC pause with the map key type string* 

go test -timeout 30m  -bench=.  -benchmem -memprofile memprofile.out -cpuprofile profile.out  -benchtime=2s -count=3
With a map of strings, GC took: 8.35775ms\ngoos: darwin
goarch: arm64
pkg: github.com/amitiwary999/go-cache
BenchmarkBigCache-10    	       1	2513947625 ns/op	206583752 B/op	 5507991 allocs/op
BenchmarkBigCache-10    	With a map of strings, GC took: 13.244958ms\n       1	2648101667 ns/op	206521816 B/op	 5507817 allocs/op
BenchmarkBigCache-10    	With a map of strings, GC took: 12.969667ms\n       1	2680075792 ns/op	206567632 B/op	 5508013 allocs/op
PASS
ok  	github.com/amitiwary999/go-cache	7.991s

*Tracking GC pause with map key type int* 

go test -timeout 30m  -bench=.  -benchmem -memprofile memprofile.out -cpuprofile profile.out  -benchtime=2s -count=3
With a map of strings, GC took: 1.404083ms\ngoos: darwin
goarch: arm64
pkg: github.com/amitiwary999/go-cache
BenchmarkBigCache-10    	       1	2739409458 ns/op	168656952 B/op	 5578545 allocs/op
BenchmarkBigCache-10    	With a map of strings, GC took: 1.817875ms\n       1	2623270083 ns/op	168584712 B/op	 5578105 allocs/op
BenchmarkBigCache-10    	With a map of strings, GC took: 2.818375ms\n       1	2663439166 ns/op	168562432 B/op	 5578132 allocs/op
PASS
