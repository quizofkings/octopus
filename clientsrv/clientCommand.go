package clientsrv

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/quizofkings/octopus/config"
	"github.com/quizofkings/octopus/network"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

//ClientCommand interface
type ClientCommand interface {
	Join(conn net.Conn)
}

type clientInfo struct {
	clients map[string]*client
	rwmutex *sync.RWMutex
	gate    network.NetCommands
}

//New create new client storage
func New() ClientCommand {
	return &clientInfo{
		clients: map[string]*client{},
		gate:    network.New(),
		rwmutex: &sync.RWMutex{},
	}
}

//Join join new client
func (c *clientInfo) Join(conn net.Conn) {

	// handler
	uniqueID := uuid.Must(uuid.NewV4()).String()
	client := newClient(conn, uniqueID)
	c.rwmutex.Lock()
	c.clients[uniqueID] = client
	c.rwmutex.Unlock()

	// logger
	logrus.Infof("join new client connection, count:%d uuid:%s", len(c.clients), uniqueID)

	// goroutine receive and health check
	go c.receiveChan(uniqueID)
	go c.healthCheck(uniqueID)
}

func (c *clientInfo) receiveChan(uniqueID string) {
	for {
		// check key exist
		cl := c.getClient(uniqueID)
		if cl == nil {
			return
		}

		// select
		select {
		case msg := <-cl.incoming:
			c.receiveCmd(c.getClient(uniqueID), msg)
		case <-c.getClient(uniqueID).disconnect:
			return
		}
	}
}

func (c *clientInfo) healthCheck(uniqueID string) {
	for {
		select {
		case uniqueID := <-c.getClient(uniqueID).disconnect:
			if c.getClient(uniqueID) != nil {
				close(c.getClient(uniqueID).disconnect)

				// delete from clients
				c.rwmutex.Lock()
				delete(c.clients, uniqueID)
				c.rwmutex.Unlock()
			}
			logrus.Warnln(fmt.Sprintf("client disconnected count:%d, uuid:%s", len(c.clients), uniqueID))
			return
		}
	}
}

//receiveCmd handle received resp command
func (c *clientInfo) receiveCmd(client *client, msg []byte) {

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

	nodeResp, err := c.gate.Write(clusterIndex, msg)
	if err != nil {
		return
	}

	// write response
	client.outgoing <- nodeResp
}

func (c *clientInfo) getClient(uniqueID string) *client {

	var (
		cl    *client
		exist bool
	)

	c.rwmutex.Lock() // R
	defer c.rwmutex.Unlock()

	if cl, exist = c.clients[uniqueID]; !exist {
		return nil
	}

	return cl
}
