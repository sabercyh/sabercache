// example.go file
// 运行前，你需要在本地启动Etcd实例，作为服务中心。

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"sabercache_client/client"
	"strconv"
	"strings"
)

var c *client.Client

func main() {
	var tcpAddr string
	flag.StringVar(&tcpAddr, "tcpAddr", "127.0.0.1:20011", "tcp地址")
	// 模拟MySQL数据库 用于peanutcache从数据源获取值
	flag.Parse()
	listen, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		fmt.Println("listen failed , err :", err)
		return
	}
	c = client.NewClient()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accepted failed , err:", err)
			continue
		}
		go process(conn)
	}
}
func process(conn net.Conn) {
	var res []byte
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		cmd := strings.Split(string(buf[:n]), " ")
		switch {
		case cmd[0] == "get":
			res = Get(cmd[1])
		case cmd[0] == "getall":
			res = GetAll()
		case cmd[0] == "set" && len(cmd) == 3:
			if Set(cmd[1], []byte(cmd[2]), -2) {
				res = []byte("true")
			} else {
				res = []byte("false")
			}
		case cmd[0] == "set" && len(cmd) == 4:
			ttl, err := strconv.Atoi(cmd[2])
			if err != nil {
				log.Println(err)
				res = []byte("err!")
			}
			if Set(cmd[1], []byte(cmd[2]), int64(ttl)) {
				res = []byte("true")
			} else {
				res = []byte("false")
			}
		case cmd[0] == "ttl":
			res = []byte(fmt.Sprint(TTL(cmd[1])))
		}
		fmt.Println(string(res))
		conn.Write(res)
	}
}

func Get(key string) (value []byte) {
	value, err := c.Get(key)
	if err != nil {
		log.Println(err)
		return []byte("err!")
	}
	return
}
func GetAll() []byte {
	var str string
	KeyValue, err := c.GetAll()
	if err != nil {
		log.Println(err)
		return []byte("err!")
	}
	for _, kv := range KeyValue {
		str += fmt.Sprintf(kv.Key + ":" + string(kv.Value))
	}
	return []byte(str)
}
func Set(key string, value []byte, ttl int64) (ok bool) {
	ok, err := c.Set(key, value, ttl)
	if !ok && err != nil {
		log.Println(err)
		return
	}
	return
}
func TTL(key string) int64 {
	ttl, err := c.TTL(key)
	if err != nil {
		log.Println(err)
		return ttl
	}
	return ttl
}
