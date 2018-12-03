package goroutinepool

import (
	"sync"
	"testing"
	"time"
)

const (
	RunTimes      = 1000000
	benchParam    = 10
	benchAntsSize = 200000
)

func demoFunc() {
	n := 10
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func demoPoolFunc(args interface{}) {
	//m := args.(int)
	//var n int
	//for i := 0; i < m; i++ {
	//	n += i
	//}
	//return nil
	n := args.(int)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func BenchmarkGoroutineWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoPoolFunc(benchParam)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkSemaphoreWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	sema := make(chan struct{}, benchAntsSize)

	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoPoolFunc(benchParam)
				<-sema
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPoolWithFunc(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := NewPool(benchAntsSize, func(i interface{}) {
		demoPoolFunc(i)
		wg.Done()
	})
	defer p.Release()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			p.Serve(benchParam)
		}
		wg.Wait()
		//b.Logf("running goroutines: %d", p.Running())
	}
	b.StopTimer()
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			go demoPoolFunc(benchParam)
		}
	}
}

func BenchmarkSemaphore(b *testing.B) {
	sema := make(chan struct{}, benchAntsSize)
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			sema <- struct{}{}
			go func() {
				demoPoolFunc(benchParam)
				<-sema
			}()
		}
	}
}

func BenchmarkAntsPool(b *testing.B) {
	p, _ := NewPool(benchAntsSize, demoPoolFunc)
	defer p.Release()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			p.Serve(benchParam)
		}
	}
	b.StopTimer()
}
