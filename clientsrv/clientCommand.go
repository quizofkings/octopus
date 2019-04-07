package clientsrv

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/quizofkings/octopus/config"
	"github.com/quizofkings/octopus/network"
	"github.com/sirupsen/logrus"
)

//ClientCommand interface
type ClientCommand interface {
	Join(conn net.Conn)
}

type clientInfo struct {
	clients map[int64]*client

	gate network.NetCommands
}

//New create new client storage
func New() ClientCommand {
	return &clientInfo{
		clients: map[int64]*client{},
		gate:    network.New(),
	}
}

//Join join new client
func (c *clientInfo) Join(conn net.Conn) {

	// logger
	logrus.Infof("join new client connection, count:%d", len(c.clients)+1)

	// handler
	uniqueID := time.Now().Unix()
	client := newClient(conn, uniqueID)
	c.clients[uniqueID] = client

	go c.receiveChan(uniqueID)
	go c.healthCheck(uniqueID)
}

func (c *clientInfo) receiveChan(uniqueID int64) {
	for {
		select {
		case msg := <-c.clients[uniqueID].incoming:
			c.receiveCmd(c.clients[uniqueID], msg)
		}
	}
}

func (c *clientInfo) healthCheck(uniqueID int64) {
	for {
		select {
		case uniqueID := <-c.clients[uniqueID].disconnect:
			logrus.Warnln(fmt.Sprintf("client disconnected #%d", uniqueID))
			close(c.clients[uniqueID].disconnect)
			delete(c.clients, uniqueID)
			return
		}
	}
}

//receiveCmd handle received resp command
func (c *clientInfo) receiveCmd(client *client, msg []byte) {

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

	nodeResp, err := c.gate.Write(clusterIndex, msg)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	// write response
	client.outgoing <- nodeResp
}
