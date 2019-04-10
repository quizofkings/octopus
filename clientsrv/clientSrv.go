package clientsrv

import (
	"sync"

	"github.com/quizofkings/octopus/network"
	"github.com/sirupsen/logrus"
)

var (
	once        sync.Once
	networkGate network.NetCommands
)

//Load client service
func Load() {
	once.Do(func() {
		logrus.Infoln("start create cluster pool...")
		networkGate = network.New()
	})
}
