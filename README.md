# go-cache

Benchmark:

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