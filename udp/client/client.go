package client

import (
	"fmt"
	"net"
)

var (
	// Connection *net.UDPConn
	Connection []*net.UDPConn
)

// Client 创建一个 UDP 连接
func Client() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:3017")

	if err != nil {
		fmt.Println("Can't resolve address: ", err)

		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		fmt.Println("Can't dial: ", err)

		panic(err)
	}

	Connection = append(Connection, conn)
}

// WriteTo 像传入参数 conn 写数据
func WriteTo(conn *net.UDPConn) {
	_, err := conn.Write([]byte("hello from the other site"))

	if err != nil {
		fmt.Println("failed:", err)
	}

	data := make([]byte, 1024)
	_, err = conn.Read(data)

	if err != nil {
		fmt.Println("failed to read UDP msg because of ", err)
	}
}
