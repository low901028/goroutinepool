package coroutinepool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	_   = 1 << (10 * iota)
	KiB  // 1024
	MiB  // 1048576
	GiB  // 1073741824
	TiB  // 1099511627776             (超过了int32的范围)
	PiB  // 1125899906842624
	EiB  // 1152921504606846976
	ZiB  // 1180591620717411303424    (超过了int64的范围)
	YiB  // 1208925819614629174706176
)

const (
	Param    = 100
	AntsSize = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

func TestAntsPoolWithFunc(t *testing.T) {
	var wg sync.WaitGroup
	p, _ := NewPool(AntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	})
	defer p.Release()

	for i := 0; i < n; i++ {
		wg.Add(1)
		p.Serve(Param)
	}
	wg.Wait()
	t.Logf("pool with func, running workers number:%d", p.Running())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}
func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestAntsPool(t *testing.T) {
	defer Release()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("pool, capacity:%d", Cap())
	t.Logf("pool, running workers number:%d", Running())
	t.Logf("pool, free workers number:%d", Free())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestCodeCov(t *testing.T) {
	_, err := NewTimingPool(-1, -1,nil)
	t.Log(err)
	_, err = NewTimingPool(1, -1, nil)
	t.Log(err)
	_, err = NewTimingPool(-1, -1, demoPoolFunc)
	t.Log(err)
	_, err = NewTimingPool(1, -1, demoPoolFunc)
	t.Log(err)

	p0, _ := NewPool(AntsSize, nil)
	defer p0.Submit(demoFunc)
	defer p0.Release()
	for i := 0; i < n; i++ {
		p0.Submit(demoFunc)
	}
	t.Logf("pool, capacity:%d", p0.Cap())
	t.Logf("pool, running workers number:%d", p0.Running())
	t.Logf("pool, free workers number:%d", p0.Free())
	p0.ReSize(AntsSize)
	p0.ReSize(AntsSize / 2)
	t.Logf("pool, after resize, capacity:%d, running:%d", p0.Cap(), p0.Running())

	p, _ := NewPool(TestSize, demoPoolFunc)
	defer p.Serve(Param)
	defer p.Release()
	for i := 0; i < n; i++ {
		p.Serve(Param)
	}
	time.Sleep(DefaultCleanIntervalTime * time.Second)
	t.Logf("pool with func, capacity:%d", p.Cap())
	t.Logf("pool with func, running workers number:%d", p.Running())
	t.Logf("pool with func, free workers number:%d", p.Free())
	p.ReSize(TestSize)
	p.ReSize(AntsSize)
	t.Logf("pool with func, after resize, capacity:%d, running:%d", p.Cap(), p.Running())
}
