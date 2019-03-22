package config

import (
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

//ClusterInfo struct
type ClusterInfo struct {
	IsMain   bool
	Name     string
	Nodes    []string
	Prefixes []string
}

//PoolConfig struct
type PoolConfig struct {
	MaxCap  int
	InitCap int
}

//Map struct
type Map struct {
	Clusters []ClusterInfo
	Prefixes map[string]int
	Port     int

	Pool      PoolConfig
	MainIndex int
}

//Reader all loaded config
var Reader = Map{
	Prefixes: map[string]int{},
}

//Opt config option struct
type Opt struct {
	Type      string // json, etcd, consul
	File      string
	Path      string
	RemoteKey string
	Host      string
	EtcdAddr  string
	ConsulKey string
}

//Load load config
func Load(opt *Opt) {

	// logger
	logrus.Infoln("load config information")

	// load
	switch opt.Type {
	default:
		viper.SetConfigName(opt.File)
		viper.AddConfigPath(opt.Path)
		err := viper.ReadInConfig()
		if err != nil {
			logrus.Fatalln(err)
		}
	case "etcd":
		viper.AddRemoteProvider("etcd", opt.Host, opt.EtcdAddr)
		viper.SetConfigType("json")
		err := viper.ReadRemoteConfig()
		if err != nil {
			logrus.Fatalln(err)
		}
	case "consul":
		viper.AddRemoteProvider("consul", opt.Host, opt.ConsulKey)
		viper.SetConfigType("json")
		err := viper.ReadRemoteConfig()
		if err != nil {
			logrus.Fatalln(err)
		}
	}

	// unmarshal config into struct
	viper.Unmarshal(&Reader)
	for i, k := range Reader.Clusters {

		// check cluster main state
		if k.IsMain {
			Reader.MainIndex = i
		}

		// group prefixes
		for _, p := range k.Prefixes {
			Reader.Prefixes[p] = i
		}
	}
}
