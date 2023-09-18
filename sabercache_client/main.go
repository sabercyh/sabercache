// example.go file
// 运行前，你需要在本地启动Etcd实例，作为服务中心。
// CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sabercache_client/client"
	"sabercache_client/util"
	"strconv"
	"strings"
)

var c *client.Client

func main() {
	util.InitConst()
	listen, err := net.Listen("tcp", util.TCPAddr)
	if err != nil {
		fmt.Println("listen failed , err :", err)
		return
	}
	log.Println("listened at : ", util.TCPAddr)
	c = client.NewClient()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accepted failed , err:", err)
			continue
		}
		log.Println("accepted conn : ", util.TCPAddr)
		go process(conn)
	}
}
func process(conn net.Conn) {
	var resp []byte
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
		case cmd[0] == "get" && len(cmd) != 1:
			resp = Get(cmd[1])
		case cmd[0] == "getall":
			resp = GetAll()
		case cmd[0] == "set" && len(cmd) == 3:
			if Set(cmd[1], []byte(cmd[2]), -1) {
				resp = []byte("true")
			} else {
				resp = []byte("false")
			}
		case cmd[0] == "set" && len(cmd) == 4:
			ttl, err := strconv.Atoi(cmd[2])
			if err != nil {
				log.Println(err)
				resp = []byte("err!")
			}
			if Set(cmd[1], []byte(cmd[3]), int64(ttl)) {
				resp = []byte("true")
			} else {
				resp = []byte("false")
			}
		case cmd[0] == "ttl" && len(cmd) != 1:
			resp = []byte(fmt.Sprint(TTL(cmd[1])))
		case cmd[0] == "exit" && len(cmd) != 1:
			break
		default:
			resp = []byte("Please enter the command in the correct format!")
		}
		conn.Write(resp)
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
		str += fmt.Sprintf(kv.Key + " : " + string(kv.Value) + "\n")
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
