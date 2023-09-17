package main

import (
	"flag"
	"fmt"
	"log"

	"sabercache_server"
)

func main() {
	var rpcAddr string
	flag.StringVar(&rpcAddr, "rpcAddr", "127.0.0.1:20001", "rpc地址")
	flag.Parse()
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	// 新建cache实例
	sc := sabercache_server.NewSaberCache(2<<10, "fifo", 30, sabercache_server.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	// New一个服务实例
	svr, err := sabercache_server.NewServer(rpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	sc.RegisterSvr(svr)
	log.Println("sabercache is running at", rpcAddr)
	// Start将不会return 除非服务stop或者抛出error
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
