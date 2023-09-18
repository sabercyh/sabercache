package main

//CGO_ENABLED=0  GOOS=linux  GOARCH=amd64  go build main.go
import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	var tcpAddr string
	flag.StringVar(&tcpAddr, "tcpAddr", "0.0.0.0:1001", "tcp地址")
	flag.Parse()
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println("dial err:", err)
		return
	}
	defer conn.Close()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, _ := inputReader.ReadString('\n')
		inputInfo := strings.Trim(input, "\n")
		_, err := conn.Write([]byte(inputInfo))
		if err != nil {
			return
		}
		if strings.ToUpper(inputInfo) == "EXIT" {
			return
		}
		buf := [512]byte{}
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("recv failed , err:", err)
		}
		fmt.Println(string(buf[:n]))
	}
}
