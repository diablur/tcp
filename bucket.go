package tcp

import (
	"sync"
	"time"
)

type ConnBucket struct {
	m  map[string]*TCPConnection
	mu *sync.RWMutex
}

func NewConnBucket() *ConnBucket {
	cb := &ConnBucket{
		m:  make(map[string]*TCPConnection),
		mu: new(sync.RWMutex),
	}
	return cb
}

func (b *ConnBucket) Put(id string, c *TCPConnection) {
	b.mu.Lock()
	b.m[id] = c
	b.mu.Unlock()
}

func (b *ConnBucket) Get(id string) *TCPConnection {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if conn, ok := b.m[id]; ok {
		return conn
	}
	return nil
}

func (b *ConnBucket) Delete(id string) {
	b.mu.Lock()
	delete(b.m, id)
	b.mu.Unlock()
}

func (b *ConnBucket) GetAll() map[string]*TCPConnection {
	b.mu.RLock()
	defer b.mu.RUnlock()
	m := make(map[string]*TCPConnection, len(b.m))
	for k, v := range b.m {
		m[k] = v
	}
	return m
}

func (b *ConnBucket) removeClosedTCPConnLoop() {
	for {
		select {
		case key := <-removeChan:
			b.Delete(key)
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}
