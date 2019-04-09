package clientsrv

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

type client struct {
	uniqueID   string
	incoming   chan []byte
	outgoing   chan []byte
	disconnect chan string
	reader     *respreader.RESPReader
	writer     *bufio.Writer
}

//newClient create new client
func newClient(connection net.Conn, uniqueID string) *client {

	// reader-writer
	writer := bufio.NewWriter(connection)
	reader := respreader.NewReader(connection)

	clientInf := &client{
		incoming:   make(chan []byte),
		outgoing:   make(chan []byte),
		disconnect: make(chan string),
		reader:     reader,
		writer:     writer,
		uniqueID:   uniqueID,
	}

	// start listening (read/write)
	clientInf.listen()

	// pinger
	// go clientInf.healthPing(connection)

	return clientInf
}

func (c *client) healthPing(connection net.Conn) {
	ping := time.NewTimer(pingInterval)
	defer ping.Stop()
	for {
		select {
		case <-ping.C:
			tmp := make([]byte, 1024)
			_, err := connection.Read(tmp)
			if err == io.EOF {
				c.flush()
				c.disconnect <- c.uniqueID
				return
			}
			ping.Reset(pingInterval)
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
	// for {
	// 	tmp := make([]byte, 512)
	// 	n, err := c.reader.Read(tmp)
	// 	if err == io.EOF {
	// 		return
	// 	}
	// 	fmt.Println("RECEIVED", tmp[:n])
	// 	c.incoming <- tmp[:n]
	// }

	// for {
	// 	fmt.Println("client.Read")
	// 	rec, err := c.reader.ReadObject()
	// 	if err == io.EOF {
	// 		return
	// 	}
	// 	c.incoming <- rec
	// }

	for {
		rec, err := c.reader.ReadObject()
		fmt.Println(err)
		if err == io.EOF {
			return
		}
		fmt.Println(string(rec), rec)
		c.incoming <- rec
	}
}

func (c *client) Write() {
	for data := range c.outgoing {
		fmt.Println("client.Write")
		_, err := c.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
		}
		c.writer.Flush()
	}
}
