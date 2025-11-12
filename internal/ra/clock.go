package ra

import "sync"

type lamport struct {
	mu  sync.Mutex
	t   int64
	id  string
	req int64 
}

func (l *lamport) tick() int64 {
	l.mu.Lock()
	l.t++
	t := l.t
	l.mu.Unlock()
	return t
}

func (l *lamport) merge(other int64) int64 {
	l.mu.Lock()
	if other > l.t {
		l.t = other
	}
	l.t++
	t := l.t
	l.mu.Unlock()
	return t
}

func (l *lamport) now() int64 {
	l.mu.Lock()
	t := l.t
	l.mu.Unlock()
	return t
}
