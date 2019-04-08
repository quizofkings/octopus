package network

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

type node struct {
	incoming chan []byte
	outgoing chan []byte
	reader   *bufio.Reader
	writer   *bufio.Writer
}

//newNode create new node
func newNode(conn net.Conn) *node {
	logrus.Infoln("new node")
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	nodeInf := &node{
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
		reader:   reader,
		writer:   writer,
	}

	// start listening (read/write)
	nodeInf.listen()

	return nodeInf
}

func (c *node) listen() {
	go c.Read()
	go c.Write()
}

//Read method
func (c *node) Read() {
	for {
		fmt.Println("readdreadread")

		bufc := respreader.NewReader(c.reader)
		bufNode, err := bufc.ReadObject()
		if err == io.EOF {
			return
		}
		c.incoming <- bufNode
	}
}

//Write method
func (c *node) Write() {
	for data := range c.outgoing {
		_, err := c.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
		}
		c.writer.Flush()
	}
}
