package network

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"

	"github.com/quizofkings/octopus/config"
	"github.com/quizofkings/octopus/respreader"
	"github.com/sirupsen/logrus"
)

const (
	maxMoved = 5
)

//NetCommands network interface
type NetCommands interface {
	Write(index int, msg []byte) ([]byte, error)
	AddNode(node string) error
}

//ClusterPool struct
type ClusterPool struct {
	conns  map[string]pool // addr => pool
	reconn chan net.Conn
}

//New create network ^-^
func New() NetCommands {

	// check nodes count
	if len(config.Reader.Clusters) == 0 {
		logrus.Fatalln("cluster info is empty!")
	}

	// logger
	logrus.Infoln("create node(s) connection")

	var clusterPoolMap = ClusterPool{
		conns:  map[string]pool{},
		reconn: make(chan net.Conn),
	}

	// do
	for _, cluster := range config.Reader.Clusters {
		for _, node := range cluster.Nodes {
			clusterPoolMap.AddNode(node)
		}
	}

	// reconnect channel
	go clusterPoolMap.reconnectHandler()

	return &clusterPoolMap
}

//reconnectHandler reconnect handler
func (c *ClusterPool) reconnectHandler() {
	for {
		select {
		case conn := <-c.reconn:
			var err error
			conn, err = net.Dial("tcp", conn.RemoteAddr().String())
			if err != nil {
				logrus.Errorln(err)
			}
		}
	}
}

//AddNode add new node when redis ASK/MOVED/initialize
func (c *ClusterPool) AddNode(node string) error {

	// check exist
	if _, exist := c.conns[node]; exist {
		return nil
	}

	p, err := newChannelPool(config.Reader.Pool.InitCap, config.Reader.Pool.MaxCap, func() (net.Conn, error) {
		logrus.Infof("create new connection, remoteAddr:%s", node)
		return net.Dial("tcp", node)
	})
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Infoln(node, "created")
	c.conns[node] = p

	return nil
}

//Write write message into connection
func (c *ClusterPool) Write(index int, msg []byte) ([]byte, error) {

	// get cluster node connection from pool
	addr, err := c.getRandomNode(index)
	if err != nil {
		return nil, err
	}

	return c.writeAction(addr, msg)
}

func (c *ClusterPool) writeAction(addr string, msg []byte) ([]byte, error) {

	// variable
	bufNode := []byte{}

	fmt.Println(addr)
	// get connection
	conn, err := c.conns[addr].get()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	fmt.Println("aaaa")

	// write
	conn.(*poolConn).n.outgoing <- msg

	fmt.Println("bbbbb")

	// read
	select {
	case saeed := <-conn.(*poolConn).n.incoming:
		bufNode = saeed
	}

	fmt.Println("ccccc", string(bufNode))

	// // write into connection
	// if _, err := conn.Write(msg); err != nil {
	// 	logrus.Errorln(err)

	// 	// close the underlying connection instead of returning it to pool
	// 	if pc, ok := conn.(*poolConn); ok {
	// 		pc.markUnusable()
	// 		pc.Close()
	// 	}

	// 	return nil, err
	// }

	// bufNode := []byte{}
	// bufChan := make(chan []byte)
	// go c.readTest(conn, bufChan)
	// select {
	// case x := <-bufChan:
	// 	fmt.Println("pashidam")
	// 	bufNode = x
	// }

	// fmt.Println(string(bufNode))

	// // reader */*~
	// bufc := respreader.NewReader(conn)
	// bufNode, err := bufc.ReadObject()
	// if err != nil {
	// 	logrus.Errorln(err)
	// 	if err == io.EOF {
	// 		logrus.Warnf("connection has been lost, remoteAddr:%s", conn.RemoteAddr().String())
	// 		// send to channel for reconnect
	// 		c.reconn <- conn
	// 	}
	// 	return nil, err
	// }

	// bufNode = []byte("+OK")

	// check moved or ask
	moved, ask, addr := redisHasMovedError(bufNode)
	if moved || ask {
		if err := c.AddNode(addr); err != nil {
			logrus.Errorln(err)
			return nil, err
		}

		return c.writeAction(addr, msg)
	}

	return bufNode, nil
}

func (c *ClusterPool) readTest(conn net.Conn, result chan []byte) {
	bufc := respreader.NewReader(conn)
	bufNode, err := bufc.ReadObject()
	if err != nil {
		logrus.Errorln(err)
		if err == io.EOF {
			logrus.Warnf("connection has been lost, remoteAddr:%s", conn.RemoteAddr().String())
			// send to channel for reconnect
			c.reconn <- conn
		}
	}

	result <- bufNode
}

func (c *ClusterPool) getRandomNode(clusterIndex int) (string, error) {

	// check requested index
	if clusterIndex > len(config.Reader.Clusters)-1 {
		return "", errors.New("cluster index bigger than registered clusters")
	}

	// get from config
	clusterNodes := config.Reader.Clusters[clusterIndex].Nodes
	lnClusterNodes := len(clusterNodes)
	var choosedNode string
	if lnClusterNodes > 1 {
		// choose randomly
		choosedNode = clusterNodes[rand.Intn(lnClusterNodes)]
	} else {
		choosedNode = clusterNodes[0]
	}

	return choosedNode, nil
}
