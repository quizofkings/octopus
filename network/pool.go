package network

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

//PoolInterface interface
type PoolInterface interface {
	// Close closes the pool and all its connections
	Close()

	// Get returns a new connection from the pool
	Get() (*Node, error)

	// Put back node into connection queue channel
	Put(node *Node) error
}

//Pool struct
type Pool struct {
	// implement
	PoolInterface

	// rw lock
	mu sync.RWMutex

	// connection channel queue
	conns chan *Node

	// maximum channel cap
	maxCap int

	// initialize channel cap
	initCap int

	// active connection count
	activeConn int

	// connection generator
	factory Factory
}

var (
	// errors
	errInvalidCapactiry = errors.New("invalid capacity settings")
	errConnNilRejecting = errors.New("connection is nil. rejecting")

	// str
	errStrFactoryFill = "factory is not able to fill the pool:"
)

//Factory type of factory method
type Factory func() (net.Conn, error)

// NewOctoPool create new octopool
// Octopool is pool of net.Conn with init/max capacity. If there is no new connection
// available in the pool, a new connection will be created via the Factory()
func NewOctoPool(initCap, maxCap int, factory Factory) (*Pool, error) {

	// check init/max capacity
	if initCap < 0 || maxCap <= 0 || initCap > maxCap {
		return nil, errInvalidCapactiry
	}

	// pool initialize
	p := &Pool{
		conns:   make(chan *Node, maxCap),
		factory: factory,
		maxCap:  maxCap,
		initCap: initCap,
	}

	// create initial connections
	for i := 0; i < initCap; i++ {

		// call factory
		node, err := p.createNode()
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("%s %s", errStrFactoryFill, err)
		}

		// add node to queue channel
		p.conns <- node
	}

	return p, nil
}

//Get node connection from queue channel
func (p *Pool) Get() (*Node, error) {

	// read lock
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case nodeConn := <-p.conns:
		return nodeConn, nil
	default:
		return p.createNode()
	}
}

//Put back node into connection queue channel
func (p *Pool) Put(node *Node) error {

	fmt.Println("CONNS:", len(p.conns))

	// check node connection
	if node.Conn == nil {
		return errConnNilRejecting
	}

	if p.conns == nil {
		// pool is closed, close passed connection
		return node.Conn.Close()
	}

	// write lock
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.conns <- node:
		return nil
	default:
		// pool is full, close passed connection
		logrus.Infoln("pool is full, close passed connection")
		return node.Conn.Close()
	}
}

//createNode create new node connection
func (p *Pool) createNode() (*Node, error) {

	// call factory
	conn, err := p.factory()
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("%s %s", errStrFactoryFill, err)
	}

	return NewNode(&conn, p), nil
}

//Close pool
func (p *Pool) Close() {

	// clear values
	p.mu.Lock()
	p.factory = nil
	p.mu.Unlock()

	// get node from queue channel and close that
	for node := range p.conns {
		node.Conn.Close()
	}
}
