package workspace

import (
	"sync"
)

// LockManager manages fine-grained path-level locks to prevent file edit races.
type LockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewLockManager constructs a LockManager instance.
func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*sync.Mutex),
	}
}

// Lock acquires a lock for a specific path.
func (lm *LockManager) Lock(path string) UnlockFunc {
	lm.mu.Lock()
	l, ok := lm.locks[path]
	if !ok {
		l = &sync.Mutex{}
		lm.locks[path] = l
	}
	lm.mu.Unlock()

	l.Lock()
	return func() {
		l.Unlock()
	}
}
