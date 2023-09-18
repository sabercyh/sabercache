package util

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	DefaultEtcdConfig = clientv3.Config{}
	RPCAddr           string
	CacheStrategy     string
)

func InitConst() {
	viper.SetConfigFile("../conf/conf.yaml") // 指定配置文件路径
	err := viper.ReadInConfig()              // 读取配置信息
	if err != nil {                          // 读取配置信息失败
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 监控配置文件变化
	viper.WatchConfig()
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{viper.GetString("ETCDEndpoints")},
		DialTimeout: time.Duration(viper.GetInt("ETCDDialTimeout")) * time.Second,
	}
	RPCAddr = viper.GetString("RPCAddr")
	CacheStrategy = viper.GetString("CacheStrategy")
	fmt.Println(RPCAddr, CacheStrategy, DefaultEtcdConfig)
}
