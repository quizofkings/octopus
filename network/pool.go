package network

import (
	"fmt"
	"net"
	"sync"
)

//poolConn struct
type poolConn struct {
	net.Conn
	mu       sync.RWMutex
	c        *channelPool
	n        *node
	unusable bool
}

func (p *poolConn) Close() error {
	fmt.Println("noooooooooo")
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
