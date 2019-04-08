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
	disconnect chan string
	reader     *bufio.Reader
	writer     *bufio.Writer
}

//newClient create new client
func newClient(connection net.Conn, uniqueID string) *client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)

	clientInf := &client{
		incoming:   make(chan []byte),
		outgoing:   make(chan []byte),
		disconnect: make(chan string),
		reader:     reader,
		writer:     writer,
	}

	// start listening (read/write)
	clientInf.listen()

	// pinger
	go clientInf.healthPing(connection, uniqueID)

	return clientInf
}

func (c *client) healthPing(connection net.Conn, uniqueID string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tmp := make([]byte, 128)
			_, err := connection.Read(tmp)
			if err == io.EOF {
				c.flush()
				c.disconnect <- uniqueID
				return
			}
		}
	}
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
		tmp := make([]byte, 1024)
		_, err := c.reader.Read(tmp)
		if err == io.EOF {
			return
		}
		c.incoming <- tmp
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
