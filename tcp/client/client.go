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
	"fmt"
	"net"

	"redalert/protocol"
)

const (
	bufferSize = 65336
)

type (
	// Conf config of server and file info
	Conf struct {
		Address  string
		Port     string
		FileName string
		PackSize int
	}

	// Client - TCP client
	Client struct {
		conf   *Conf
		conn   *net.TCPConn
		proto  *protocol.Proto
		handle handler
	}
)

// NewClient creat a new tcp client
func NewClient(conf *Conf, hand handler) (*Client, error) {
	addr, err := net.ResolveTCPAddr("tcp", conf.Address+":"+conf.Port)
	if err != nil {
		fmt.Printf("[ERROR] Can't resolve tcp address with error %v \n", err)
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("[ERROR] Dial crash with error: %v \n", err)
		return nil, err
	}

	client := &Client{
		conf:   conf,
		conn:   conn,
		proto:  &protocol.Proto{},
		handle: hand,
	}

	client.prepareBuffer()

	return client, nil
}
func (c *Client) prepareBuffer() {
	c.conn.SetReadBuffer(bufferSize)
	c.conn.SetWriteBuffer(bufferSize)
}
