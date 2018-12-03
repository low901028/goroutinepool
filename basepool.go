package coroutinepool

import (
	"math"
	"errors"
)

const (
	// DefaultAntsPoolSize is the default capacity for a default goroutine pool.
	DefaultAntsPoolSize = math.MaxInt32

	// DefaultCleanIntervalTime is the interval time to clean up goroutines.
	DefaultCleanIntervalTime = 5
)


// Init a instance pool when importing ants.
var defaultPool, _ = NewPool(DefaultAntsPoolSize,nil)

// Submit submits a task to pool.
func Submit(task func()) error {
	return defaultPool.Submit(task)
}

func Serve(args interface{}) error{
	return defaultPool.Serve(args)
}

// Running returns the number of the currently running goroutines.
func Running() int {
	return defaultPool.Running()
}

// Cap returns the capacity of this default pool.
func Cap() int {
	return defaultPool.Cap()
}

// Free returns the available goroutines to work.
func Free() int {
	return defaultPool.Free()
}

// Release Closes the default pool.
func Release() {
	defaultPool.Release()
}

var (
	ErrInvalidPoolSize   = errors.New("invalid size for pool")
	ErrInvalidPoolExpiry = errors.New("invalid expiry for pool")
	ErrPoolClosed        = errors.New("this pool has been closed")
)
