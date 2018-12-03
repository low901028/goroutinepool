package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"GoDemo/coroutinepool"
)

var sum int32

func myFunc(i interface{}) error {
	n := i.(int32)
	atomic.AddInt32(&sum, n)
	fmt.Printf("run with %d\n", n)
	return nil
}

func demoFunc() error {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
	return nil
}

func main() {
	defer coroutinepool.Release()

	runTimes := 1000

	// use the common pool
	var wg sync.WaitGroup
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		coroutinepool.Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", coroutinepool.Running())
	fmt.Printf("finish all tasks.\n")

	// use the pool with a function
	// set 10 the size of goroutine pool and 1 second for expired duration
	p, _ := coroutinepool.NewPool(10, func(i interface{}) {
		myFunc(i)
		wg.Done()
	})
	defer p.Release()
	// submit tasks
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		p.Serve(int32(i))
	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks, result is %d\n", sum)
}
