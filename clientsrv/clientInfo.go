package clientsrv

import (
	"bufio"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type client struct {
	incoming   chan []byte
	outgoing   chan []byte
	disconnect chan int64
	reader     *bufio.Reader
	writer     *bufio.Writer
}

//newClient create new client
func newClient(connection net.Conn, uniqueID int64) *client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	clientInf := &client{
		incoming:   make(chan []byte),
		outgoing:   make(chan []byte),
		disconnect: make(chan int64),
		reader:     reader,
		writer:     writer,
	}

	clientInf.listen()

	// pinger
	go func(connection net.Conn, uniqueID int64, c *client) {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tmp := make([]byte, 12)
				_, err := connection.Read(tmp)
				if err == io.EOF {
					c.flush()
					c.disconnect <- uniqueID
					return
				}
			}
		}
	}(connection, uniqueID, clientInf)

	return clientInf
}

func (c *client) flush() {
	close(c.incoming)
	close(c.outgoing)
}

func (c *client) listen() {
	go c.Read()
	go c.Write()
}

func (c *client) Read() {
	for {
		tmp := make([]byte, 512)
		_, err := c.reader.Read(tmp)
		c.incoming <- tmp
		if err == io.EOF {
			return
		}
	}
}

func (c *client) Write() {
	for data := range c.outgoing {
		_, err := c.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
		}
		c.writer.Flush()
	}
}
