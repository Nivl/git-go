package syncutil

import (
	"sync"

	"github.com/gogf/gf/encoding/ghash"
)

// NamedMutex is a struct allowing to lock/unlock using a key
// It is expected that 2 keys may collide
type NamedMutex struct {
	locks []sync.RWMutex
	size  uint32
}

// NewNamedMutex creates a new NamedMutex with the given capacity.
// If the max number is below 2, 2 will be used.
// using a prime number as max offers better performance
func NewNamedMutex(maxMutexes uint32) *NamedMutex {
	if maxMutexes < 2 {
		maxMutexes = 2
	}

	return &NamedMutex{
		size:  maxMutexes,
		locks: make([]sync.RWMutex, maxMutexes),
	}
}

// Lock locks the provided key. If the lock is already in use, the
// calling goroutine blocks until the mutex is available.
func (mu *NamedMutex) Lock(key []byte) {
	mu.locks[ghash.SDBMHash(key)%mu.size].Lock()
}

// Unlock unlocks the provided key. It is a run-time error if the key
// is not locked on entry to Unlock.
func (mu *NamedMutex) Unlock(key []byte) {
	mu.locks[ghash.SDBMHash(key)%mu.size].Unlock()
}

// RLock locks rw for reading.
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
func (mu *NamedMutex) RLock(key []byte) {
	mu.locks[ghash.SDBMHash(key)%mu.size].RLock()
}

// RUnlock undoes a single RLock call; it does not affect other
// simultaneous readers. It is a run-time error if rw is not locked
// for reading on entry to RUnlock.
func (mu *NamedMutex) RUnlock(key []byte) {
	mu.locks[ghash.SDBMHash(key)%mu.size].RUnlock()
}
