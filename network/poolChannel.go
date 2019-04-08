package network

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

type pool interface {
	get() (net.Conn, error)
	Close()
	Len() int
}

type channelPool struct {
	mu        sync.RWMutex
	conns     chan net.Conn
	connCount int

	// net.Conn generator
	factory Factory
}

// Factory is a function to create new connections.
type Factory func() (net.Conn, error)

//newChannelPool create channel pool
func newChannelPool(initialCap, maxCap int, factory Factory) (pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:   make(chan net.Conn, maxCap),
		mu:      sync.RWMutex{},
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool, %s", err)
		}
		c.connCount++
		c.conns <- conn
	}

	return c, nil
}

func (c *channelPool) getConnsAndFactory() (chan net.Conn, Factory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) get() (net.Conn, error) {

	conns, _ := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	// check pool capacity
	// go func() {
	// 	c.mu.Lock()
	// 	defer c.mu.Unlock()

	// 	if cap(c.conns) >= c.connCount {
	// 		return
	// 	}

	// 	conn, err := factory()
	// 	if err != nil {
	// 		logrus.Errorln(err)
	// 		return
	// 	}
	// 	c.connCount++
	// 	c.conns <- conn
	// }()

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	fmt.Println("00")
	select {
	case conn := <-conns:
		fmt.Println("0A")
		if conn == nil {
			return nil, ErrClosed
		}

		fmt.Println("0B")
		return c.wrapConn(conn), nil
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) put(conn net.Conn) error {

	// check
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	// lock
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}

// newConn wraps a standard net.Conn to a poolConn net.Conn.
func (c *channelPool) wrapConn(conn net.Conn) net.Conn {
	p := &poolConn{
		c: c,
		n: newNode(conn),
	}
	p.Conn = conn
	return p
}
