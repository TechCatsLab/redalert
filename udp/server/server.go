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
 *     Initial: 2017/08/10        Yusan Kurban
 */

package main

import (
	"fmt"
	"net"
	"os"

	"redalert/udp/protocal"
)

var (
	close = make(chan struct{}, 1)
)

type Service struct {
	Conn *net.UDPConn
}

func NewService(port string) *Service {
	var udpPort string

	if port == "" {
		udpPort = ":" + protocal.DefaultUDPPort
	} else {
		udpPort = ":" + port
	}

	addr, err := net.ResolveUDPAddr("udp", udpPort)
	if err != nil {
		fmt.Println("Can't resolve addr:", err.Error())
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("listen error:", err.Error())
		panic(err)
	}
	defer conn.Close()

	server := &Service{
		Conn: conn,
	}

	go server.handlerClient()

	return server
}

func (c *Service) handlerClient() {
	for {
		select {
		case <-close:
			os.Exit(1)
		}
	}
}

func Close() {
	close <- struct{}{}
}
