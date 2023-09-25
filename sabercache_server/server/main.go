package main

//./main --rpcAddr 127.0.0.1:20001 --cacheStrategy lru
// go run main.go --rpcAddr 127.0.0.1:20001 --cacheStrategy lru
//CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build main.go
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sabercache_server/util"
	"strings"

	"sabercache_server"
)

func main() {
	// 新建cache实例
	sc := sabercache_server.NewSaberCache(2<<10, util.CacheStrategy, sabercache_server.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			file, err := os.OpenFile("../file/mysql.txt", os.O_RDONLY, 0777)
			if err != nil {
				return []byte{}, err
			}
			reader := bufio.NewReader(file)
			for {
				bytes, _, err := reader.ReadLine()
				if err == io.EOF {
					return []byte{}, fmt.Errorf("not found the key:%s", key)
				} else if err != nil {
					return []byte{}, err
				}
				kv := strings.Split(string(bytes), " ")
				if kv[0] == key {
					return []byte(kv[1]), nil
				}
			}
		}))
	// New一个服务实例
	svr, err := sabercache_server.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	sc.RegisterSvr(svr)
	log.Println("sabercache is running at", util.RPCAddr, "cache Strategy:", util.CacheStrategy)
	// Start将不会return 除非服务stop或者抛出error
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
