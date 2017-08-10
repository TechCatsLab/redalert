/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co.,Ltd..
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
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
 * Revision History:
 *     Initial: 2017/08/10        Liu Jiachang
 */

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
