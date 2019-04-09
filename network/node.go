package network

import (
	"bufio"
	"io"
	"net"

	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

//Node struct
type Node struct {
	net.Conn

	// parent pool
	p *Pool

	// io channel
	Incoming   chan []byte
	Outgoing   chan []byte
	WriteError chan error

	// bufio rw
	reader *respreader.RESPReader
	writer *bufio.Writer
}

//NewNode create new node
func NewNode(conn *net.Conn, pool *Pool) *Node {
	nodeInf := &Node{
		Incoming:   make(chan []byte),
		Outgoing:   make(chan []byte),
		WriteError: make(chan error),
		reader:     respreader.NewReader(*conn), // Size ^-^
		writer:     bufio.NewWriter(*conn),      // Size ^-^
		p:          pool,
		Conn:       *conn,
	}

	// start listening
	nodeInf.listen()

	return nodeInf
}

//Close close node
func (n *Node) Close() error {
	logrus.Infoln("close node, put back into connection queue channel ")
	return n.p.Put(n)
}

func (n *Node) listen() {
	go n.Read()
	go n.Write()
}

func (n *Node) flush() {
	close(n.Incoming)
	close(n.Outgoing)
}

//Read method
func (n *Node) Read() {
	for {
		rec, err := n.reader.ReadObject()
		if err == io.EOF {
			return
		}
		n.Incoming <- rec
		// n.Incoming <- []byte("+OK\n")
		// return
	}
}

//Write method
func (n *Node) Write() {
	for data := range n.Outgoing {
		_, err := n.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
			n.WriteError <- err
		}
		n.writer.Flush()
	}
}
