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
	"encoding/binary"
	"fmt"
	"log"
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
		info   *FileInfo
		close  chan struct{}
	}
)

// NewClient create a new tcp client
func NewClient(conf *Conf) *Client {
	addr, err := net.ResolveTCPAddr("tcp", conf.Address+":"+conf.Port)
	if err != nil {
		log.Fatalf("[ERROR] Can't resolve tcp address with error %v \n", err)
		return nil
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatalf("[ERROR] Dial crash with error: %v \n", err)
		return nil
	}

	client := &Client{
		conf:  conf,
		conn:  conn,
		proto: &protocol.Proto{},
		close: make(chan struct{}),
		info: &FileInfo{
			filePack:  make([]byte, conf.PackSize),
			replyPack: make([]byte, protocol.ReplySize),
		},
	}

	client.info.client = client
	client.handle = &Provider{
		client: client,
	}
	client.prepareBuffer()

	if err = client.info.initFile(conf.FileName); err != nil {
		client.handle.OnError(err)
	}

	return client
}

// Start start sending
func (c *Client) Start() {
	c.receive(c.conn)
}

func (c *Client) prepareBuffer() {
	c.conn.SetReadBuffer(bufferSize)
	c.conn.SetWriteBuffer(bufferSize)
}

func (c *Client) receive(conn *net.TCPConn) {
	go func() {
		select {
		case <-c.close:
			c.handle.OnClose()
		}
	}()

	if err := c.info.consult(); err != nil {
		c.handle.OnError(err)
	}

	for {
		_, err := conn.Read(c.info.replyPack)
		if err != nil {
			c.handle.OnError(err)
		}

		packOrder := binary.LittleEndian.Uint32(c.info.replyPack)
		fmt.Printf("[RECEIVE] pack order %d \n", packOrder)

		if packOrder == protocol.ReplyError {
			c.handle.OnError(errFromServer)
		}

		if packOrder != c.proto.PackOrder {
			c.handle.OnError(errPackOrder)
		}

		err = c.info.SendFile(c.conf.PackSize)
		if err != nil {
			c.handle.OnError(err)
		}

		if c.info.filePack[0] == protocol.HeaderFileFinishType {
			c.close <- struct{}{}
		}
	}
}
