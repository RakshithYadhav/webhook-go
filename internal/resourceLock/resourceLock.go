package resourcelock

import (
	"sync"

	"github.com/google/uuid"
)

type ResourceLock struct {
	mu    sync.Mutex
	locks map[uuid.UUID]*sync.Mutex
}

func NewResourceLock() *ResourceLock {
	return &ResourceLock{
		locks: make(map[uuid.UUID]*sync.Mutex),
	}
}

func (rl *ResourceLock) Lock(resId uuid.UUID) {
	rl.mu.Lock()
	m, ok := rl.locks[resId]
	if !ok {
		m = &sync.Mutex{}
		rl.locks[resId] = m
	}
	rl.mu.Unlock()
	m.Lock()
}

func (rl *ResourceLock) Unlock(resId uuid.UUID) {
	rl.mu.Lock()
	m := rl.locks[resId]
	rl.mu.Unlock()
	m.Unlock()
}
