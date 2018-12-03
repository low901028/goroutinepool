# goroutinepool
golang版协程池实现
### 性能压测
go test -bench=. -benchmem=true -run=none -test.memprofile mem.out
