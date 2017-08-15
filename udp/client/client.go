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
	"os"

	"redalert/udp/protocal"
)

// Client - UDP Client
type Client struct {
	conf      *Conf
	conn      *net.UDPConn
	handler   Handler
	replyByte []byte
	headBytes []byte
	fileBytes []byte
	file      *os.File
	proto     *protocal.Proto
	fileInfo  os.FileInfo
}

// NewClient 创建一个 UDP 客户端
func NewClient(conf *Conf, handler Handler) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", conf.RemoteAddress+":"+conf.RemotePort)
	if err != nil {
		fmt.Println("Can't resolve address: ", err)

		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		fmt.Println("Can't dial: ", err)

		return nil, err
	}

	client := Client{
		conf:      conf,
		conn:      conn,
		handler:   handler,
		replyByte: make([]byte, protocal.ReplySize),
		headBytes: make([]byte, protocal.FixedHeaderSize),
	}

	return &client, nil
}

// StartRun - Client start run
func (c *Client) StartRun() {

}

// WriteFirst - 第一次连接服务器，商量协议
func (c *Client) WriteFirst() error {
	c.headBytes[0] = byte(protocal.HeaderRequestType)
	protocal.PutInt16(c.headBytes[protocal.HeaderSizeOffset:], c.proto.HeaderSize)
	protocal.PutInt64(c.headBytes[protocal.FileSizeOffset:], c.proto.FileSize)
	protocal.PutInt16(c.headBytes[protocal.PackSizeOffset:], c.proto.PackSize)
	protocal.PutInt32(c.headBytes[protocal.PackCountOffset:], c.proto.PackCount)
	protocal.PutInt32(c.headBytes[protocal.PackOrderOffset:], c.proto.PackOrder)
	protocal.PutString(c.headBytes[protocal.FixedHeaderSize:], c.fileInfo.Name())

	_, err := c.conn.Write(c.headBytes)

	return err
}
