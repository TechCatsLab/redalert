package client

import (
	"bufio"
	"fmt"
	"net"
)

var (
	Connection []*net.TCPConn
	bytebuf []byte
)

func TcpClient() {
	var tcpAddr *net.TCPAddr

	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println("Resolve tcp addr error:", err)
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Dial tcp error:", err)
		panic(err)
	}

	fmt.Println(conn.LocalAddr())
	Connection = append(Connection, conn)
}

func ReadMessage(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println(msg)
	}
}

func WriteMessage(conn *net.TCPConn) {
	b := []byte("yusakkurbaebi\n")
	_, err := conn.Write(b)

	if err != nil {
		fmt.Println("client write error:", err)
		return
	}
	_, err = conn.Read(bytebuf)
	if err != nil {
		fmt.Println("client read error:", err)
		return
	}
}
