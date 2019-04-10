package network

import (
	"bufio"
	"io"
	"net"

	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

type node struct {
	incoming chan []byte
	outgoing chan []byte
	disc     chan string
	reader   *respreader.RESPReader
	writer   *bufio.Writer
}

//newNode create new node
func newNode(conn net.Conn) *node {

	// reader-writer
	writer := bufio.NewWriter(conn)
	reader := respreader.NewReader(conn)

	nodeInf := &node{
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
		disc:     make(chan string),
		reader:   reader,
		writer:   writer,
	}

	// start handling message
	// go nodeInf.handleMessage()

	// i/o start
	go nodeInf.Read()
	go nodeInf.Write()

	return nodeInf
}

//Read read msg
func (c *node) Read() {
	// for {
	msg, err := c.reader.ReadObject()
	if err == io.EOF {
		return
	}
	c.incoming <- msg
	// }
}

//Write write msg
func (c *node) Write() {
	for data := range c.outgoing {
		_, err := c.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		c.writer.Flush()
	}
}

func (c *node) Close() {
	close(c.incoming)
	close(c.outgoing)
	c.reader = nil
	c.writer = nil
}
