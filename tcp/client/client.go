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
	"io"
	"net"
)

type TcpClient struct {
	RemoteAddr *net.TCPAddr
	Conn       *net.TCPConn
	MsgQueue   chan []byte
}

func NewTcpClient(remoteAddr string) (*TcpClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		fmt.Println("Resolve tcp addr error:", err)

		return nil, err
	}

	return &TcpClient{RemoteAddr: tcpAddr, MsgQueue: make(chan []byte, 1024)}, nil
}

func (cli *TcpClient) DialTcp() error {
	conn, err := net.DialTCP("tcp", nil, cli.RemoteAddr)
	if err == nil {
		cli.Conn = conn
	}

	go func() {
		reader := bufio.NewReader(cli.Conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				cli.Reconnection()
			}

			if msg != "" {
				cli.MsgQueue <- []byte(msg)
			}
		}
	}()

	return err
}

func (cli *TcpClient) ReadMessage() {
	go func() {
		for {
			select {
			case <-cli.MsgQueue:
				// do something with the msg in channel
			}
		}
	}()
}

func (cli *TcpClient) WriteMessage(msg string) error {
	_, err := cli.Conn.Write([]byte(msg))

	return err
}

func (cli *TcpClient) Reconnection() error {
	return cli.DialTcp()
}
