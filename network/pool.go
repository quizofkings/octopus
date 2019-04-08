package network

import (
	"net"
	"sync"
)

//poolConn struct
type poolConn struct {
	net.Conn
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

func (p *poolConn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return p.c.put(p.Conn)
}

func (p *poolConn) markUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}
