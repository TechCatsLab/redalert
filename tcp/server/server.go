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
 *     Initial: 2017/08/11        Sun Anxiang
 */

package server

import (
	"net"
	"time"
	"bytes"
	"fmt"
)

type TcpServer struct {
	conf      *Conf
	listener  *net.TCPListener
	shutdown  chan struct{}
	sender    chan *Msg
	connCount int
}

type Msg struct {
	content bytes.Buffer
	remote  net.Addr
}

func NewTcpServer(conf *Conf) *TcpServer {
	return &TcpServer{
		conf:     conf,
		sender:   make(chan *Msg),
		shutdown: make(chan struct{}),
	}
}

func (server *TcpServer) Init() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server.conf.Addr+server.conf.Port)
	if err != nil {
		return err
	}

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	server.listener = tcpListener

	return nil
}

func (server *TcpServer) Start() {
	defer func() {
		server.listener.Close()
	}()

	for {
		select {
		case <-server.shutdown:
			return
		default:
			conn, err := server.listener.AcceptTCP()
			if err != nil {
				continue
			}

			go NewClient(conn, server).Start()
			fmt.Println("connection:", )

			server.connCount++
			go server.stopAcceptConn()
		}
	}
}

func (server *TcpServer) Shutdown() {
	server.shutdown <- struct{}{}
}

func (server *TcpServer) stopAcceptConn() {
	if server.connCount == server.conf.ConnLimit {
		server.listener.SetDeadline(time.Now())
		close(server.shutdown)
	}
}

func (server *TcpServer) Send(msg *Msg) {
	server.sender <- msg
}
