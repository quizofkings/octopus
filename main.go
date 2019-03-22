package main

import (
	"fmt"
	"log"
	"net"

	"github.com/quizofkings/octopus/clientsrv"
	"github.com/quizofkings/octopus/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	envType, envFile, envPath, envHost      string
	envRemoteKey, envEtcdAddr, envConsulKey string
)

var rootCmd = &cobra.Command{
	Use:   "gate",
	Short: "Octopus Gate",
	Run: func(cmd *cobra.Command, args []string) {

		// load config
		config.Load(&config.Opt{
			Type:      envType,
			File:      envFile,
			Path:      envPath,
			RemoteKey: envRemoteKey,
			Host:      envHost,
			EtcdAddr:  envEtcdAddr,
			ConsulKey: envConsulKey,
		})

		// create client storage
		clientStorage := clientsrv.New()

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Reader.Port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		for {
			// Wait for a connection.
			conn, err := l.Accept()
			if err != nil {
				logrus.Fatalln(err)
			}

			// call client storage for new connection
			go clientStorage.Join(conn)
		}

	},
}

func main() {

	// flagger
	rootCmd.PersistentFlags().StringVarP(&envType, "type", "t", "file", "config type (file, etcd, consul)")
	rootCmd.PersistentFlags().StringVarP(&envPath, "path", "p", "/", "config path (default is /)")
	rootCmd.PersistentFlags().StringVarP(&envFile, "file", "f", "config", "config file name (without extenstion)")
	rootCmd.PersistentFlags().StringVarP(&envRemoteKey, "remotekey", "r", "", "etcd/consul remote key (default is empty)")
	rootCmd.PersistentFlags().StringVar(&envHost, "host", "", "config remote host (ex. http://127.0.0.1:4001)")
	rootCmd.PersistentFlags().StringVar(&envEtcdAddr, "etcdaddr", "", "etcd file address")
	rootCmd.PersistentFlags().StringVar(&envConsulKey, "consulkey", "", "consul k/v namespace")

	// cobra console command
	rootCmd.Execute()
}
