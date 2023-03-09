package utils

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"strings"
	"sync"
)

// Semaphore is a struct wrapping a sema`phore thast makes it easier to debug deadlocvks and failures to release
type Semaphore struct {
	lock    *semaphore.Weighted
	holders map[string]int64
	mut     sync.Mutex
}

func NewSemaphore(count int64) *Semaphore {
	return &Semaphore{
		lock:    semaphore.NewWeighted(count),
		holders: make(map[string]int64),
	}
}

func (l Semaphore) Acquire(ctx context.Context, count int64, owner string) error {
	err := l.lock.Acquire(ctx, count)
	if err == nil {
		l.mut.Lock()
		l.holders[owner] = count
		l.mut.Unlock()
	}
	return err
}

func (l Semaphore) Release(i int64, owner string) {
	count, ok := l.holders[owner]
	if !ok {
		panic(fmt.Sprintf("no lock held by %s", owner))
	}
	if count < i {
		panic(fmt.Sprintf("%s lock count: %d, trying to release %d", owner, count, i))
	}
	l.mut.Lock()
	l.holders[owner] = count - i
	if l.holders[owner] == 0 {
		delete(l.holders, owner)
	}
	l.mut.Unlock()
	l.lock.Release(i)
}

func (l Semaphore) String() any {
	var sb strings.Builder
	for k, v := range l.holders {
		sb.WriteString(fmt.Sprintf("%s : %d\n", k, v))
	}
	return sb.String()
}
