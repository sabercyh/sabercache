package main

// go run main.go --rpcAddr 127.0.0.1:20001 --cacheStrategy lru
//CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build main.go
import (
	"flag"
	"fmt"
	"log"

	"sabercache_server"
)

func main() {
	var rpcAddr string
	var cacheStrategy string
	flag.StringVar(&rpcAddr, "rpcAddr", "127.0.0.1:20001", "rpc地址")
	flag.StringVar(&cacheStrategy, "cacheStrategy", "fifo", "缓存淘汰策略")
	flag.Parse()
	var mysql = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	// 新建cache实例
	sc := sabercache_server.NewSaberCache(2<<10, cacheStrategy, 30, sabercache_server.RetrieverFunc(
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
	log.Println("sabercache is running at", rpcAddr, "cache Strategy:", cacheStrategy)
	// Start将不会return 除非服务stop或者抛出error
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
