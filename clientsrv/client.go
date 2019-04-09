package clientsrv

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/quizofkings/octopus/config"
	"github.com/quizofkings/octopus/respreader"
	"github.com/quizofkings/octopus/utils"
	"github.com/sirupsen/logrus"
)

type client struct {
	uniqueID string
	incoming chan []byte
	outgoing chan []byte
	disc     chan string
	reader   *respreader.RESPReader
	writer   *bufio.Writer
}

//HandleClient create new client
func HandleClient(conn net.Conn) {

	// logger
	logrus.Infoln("new client connected to octopus!")

	// generate uniqueID
	uniqueID := utils.RandStringRunes(15)

	// reader-writer
	writer := bufio.NewWriter(conn)
	reader := respreader.NewReader(conn)

	clientInf := &client{
		incoming: make(chan []byte),
		outgoing: make(chan []byte),
		disc:     make(chan string),
		reader:   reader,
		writer:   writer,
		uniqueID: uniqueID,
	}

	// handle disconencted client
	defer func() {
		clientInf.disc <- uniqueID
	}()

	// start handling message
	go clientInf.handleMessage()

	// i/o start
	go clientInf.Read()
	clientInf.Write()
}

//Read read msg
func (c *client) Read() {
	for {
		msg, err := c.reader.ReadObject()
		if err == io.EOF {
			break
		}
		c.incoming <- msg
	}
}

//Write write msg
func (c *client) Write() {
	for data := range c.outgoing {
		_, err := c.writer.Write(data)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		c.writer.Flush()
	}
}

//handleMessage handle income message
func (c *client) handleMessage() {
	for {
		select {
		case msg := <-c.incoming:
			c.handleIncome(msg)
		case <-c.disc:
			logrus.Warnln("client has been disconnected!")
		}
	}
}

//handleIncome handle income message
func (c *client) handleIncome(msg []byte) {

	// handle close channel
	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			err := fmt.Errorf("%v", r)
			logrus.Warnf("write: error writing on channel: %v", err)
			return
		}
	}()

	// variable
	var (
		mainNode     = true
		clusterIndex int
		prefix       string
	)

	// find hash path
	for prefix, clusterIndex = range config.Reader.Prefixes {
		bufStr := string(msg)
		splits := strings.Split(bufStr, "\n")
		if len(splits) < 5 {
			break
		}

		if strings.HasPrefix(splits[4], prefix) {
			mainNode = false
			break
		}
	}

	if mainNode {
		clusterIndex = config.Reader.MainIndex
	}

	if len(msg) == 0 {
		return
	}

	// send to redis nodes
	nodeResp, err := networkGate.Write(clusterIndex, msg)
	if err != nil {
		return
	}

	// send ro outgoing channel
	c.outgoing <- nodeResp
}
