package goroutinepool

import (
	"time"
	"sync"
	"sync/atomic"
)

//type pf func(interface{})

// Pool accept the tasks from client ,it limits the total of
// goroutines to a given number by recycling goroutines.
type Pool struct {
	// capacity of the pool
	capacity int32

	// the number of the currently running goroutines
	running int32

	// the expired time of every worker （second）
	expiryDuration time.Duration

	// a slice that store the available workers
	workers []*Worker

	// notice the pool to closed itself
	release chan struct{} // 使用匿名struct

	// synchronous operation
	lock sync.Mutex

	// waiting to get a idle worker
	cond *sync.Cond

	// pf is the function for processing tasks
	poolFunc func(interface{})

	// only one execute
	once sync.Once
}

// clear expired workers periodically
func (p *Pool) periodicallyPurge(){
	heartbeat := time.NewTicker(p.expiryDuration)  // 心跳检查
	defer heartbeat.Stop()

	for range heartbeat.C{
		currentTime := time.Now()
		p.lock.Lock()
		idleWorkers := p.workers
		if len(idleWorkers) == 0 && p.running == 0 && len(p.release) > 0{  // 无空闲worker
			p.lock.Unlock()
			return
		}

		n := -1
		for i, w := range idleWorkers{
			if currentTime.Sub(w.recycleTime) <= p.expiryDuration{
				break
			}
			n = i
			w.task <- nil
			idleWorkers[i] = nil
		}
		if n > -1 {
			if n >= len(idleWorkers)-1{
				p.workers = idleWorkers[:0]
			} else{
				p.workers = idleWorkers[n+1:]
			}
		}

		p.lock.Unlock()
	}
}

//
func NewPool(size int, pf func(interface{})) (*Pool, error){
	return NewTimingPool(size, DefaultCleanIntervalTime, pf)
}

// generates an instance of pool with a specific function and a custom timed task.
func NewTimingPool(size int, expiry int, pf func(interface{}))(*Pool, error){
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}
	if expiry <= 0 {
		return nil, ErrInvalidPoolExpiry
	}

	p := &Pool{
		capacity:		int32(size),
		release:		make(chan struct{}, 1),
		expiryDuration: time.Duration(expiry) * time.Second,
		poolFunc:				pf,
	}
	p.cond = sync.NewCond(&p.lock)
	go p.periodicallyPurge()

	return p, nil
}

// submit a task to pool
func (p *Pool) Submit(task func()) error{
	if len(p.release) > 0 {
		return ErrPoolClosed
	}
	p.getWorker().task <- task
	return nil
}

func (p *Pool) Serve(args interface{}) error {
	if len(p.release) > 0 {
		return ErrPoolClosed
	}
	p.getWorker().args <- args
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns a available goroutines to work.
func (p *Pool) Free() int {
	return int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Resize change the capacity of this pool
func (p *Pool) ReSize(size int){
	if size == p.Cap(){
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
	diff := p.Running() - size
	for i := 0; i < diff; i++{
		p.getWorker().task <- nil
	}
}

// Release Closed this pool
func (p *Pool) Release() error{
	p.once.Do(func() {   // 保证释放操作只执行一次
		p.release <- struct{}{}  // 发送释放信号
		p.lock.Lock()
		idleWorkers := p.workers  // 将pool池中创建的worker 一一进行重置
		for i, w := range idleWorkers {
			w.task <- nil                    // task重置
			idleWorkers[i] = nil             // worker重置
		}
		p.workers = nil                      // 池中worker重置
		p.lock.Unlock()
	})
	return nil
}

// incRunning increases the number of the currently running goroutines.
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) getWorker() *Worker {
	var w *Worker
	waiting := false

	p.lock.Lock()
	defer p.lock.Unlock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		waiting = p.Running() >= p.Cap()
	} else {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
	}

	if waiting{
		for {
			p.cond.Wait()
			l := len(p.workers) - 1
			if l < 0 {
				continue
			}
			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			break
		}
	} else if w == nil{
		w = &Worker{
			pool: p,
			task: make(chan func(), 1),
			args: make(chan interface{}, 1),
		}
		w.run()
		p.incRunning()
	}
	return w
}

// putWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) putWorker(worker *Worker) {
	worker.recycleTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, worker)
	// notify there is available worker
	p.cond.Signal()
	p.lock.Unlock()
}
















