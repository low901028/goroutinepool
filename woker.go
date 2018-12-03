package goroutinepool

import (
	"time"
)

// Worker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type Worker struct {
	// pool who owns this worker
	pool *Pool
	// task is a job should be done
	task chan func()

	args chan interface{}

	// recycleTime will be update when putting a worker back into queue
	recycleTime time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *Worker) run(){
	go func() {
		for {
			select {
			case f := <- w.task:
				if f == nil {
					w.pool.decRunning()
					return
				}
				f()
				w.pool.putWorker(w)
			case args := <- w.args:
				if args == nil {
					w.pool.decRunning()
					return
				}
				w.pool.poolFunc(args)
				w.pool.putWorker(w)
			}
		}
	}()
}
