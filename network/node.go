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

	// unusable
	unusable bool

	// bufio rw
	reader *respreader.RESPReader
	writer *bufio.Writer
}

//NewNode create new node
func NewNode(conn *net.Conn, pool *Pool) *Node {

	// keep alive
	(*conn).(*net.TCPConn).SetKeepAlive(true)

	// node config
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
			n.unusable = true
			return
		}
		n.Incoming <- rec
	}
	// for {
	// 	fmt.Println("node.Read")
	// 	rec, err := n.reader.ReadObject()
	// 	n.Incoming <- rec
	// 	if err != nil {
	// 		// n.unusable = true
	// 		logrus.WithField("unsusable", "yes").Errorln(err)

	// 		return
	// 	}
	// 	// n.Incoming <- []byte("+OK\n")
	// }
}

//Write method
func (n *Node) Write() {
	for data := range n.Outgoing {
		if len(data) == 0 {
			return
		}
		_, err := n.writer.Write(data)
		if err != nil {
			logrus.WithField("mode", "write").Errorln(err)
			n.WriteError <- err
			return
		}
		n.writer.Flush()
	}
}
