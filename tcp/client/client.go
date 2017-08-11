/*
* MIT License
*
* Copyright (c) 2017 SmartestEE Co.,Ltd..
*
* Permission is hereby granted, free of charge, to any person obtaining a copy of
* this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
 */

/*
* Revision History
*     Initial: 2017/08/11          Yusan Kurban
 */

package client

import (
	"bufio"
	"fmt"
	"net"
)

var (
	Connection []*net.TCPConn
	bytebuf    []byte
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
