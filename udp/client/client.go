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
	"io"
	"log"
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
	file      *os.File
	proto     *protocal.Proto
	fileInfo  os.FileInfo
}

// NewClient 创建一个 UDP 客户端
func NewClient(conf *Conf, handler Handler) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", conf.RemoteAddress+":"+conf.RemotePort)
	if err != nil {
		log.Println("Can't resolve address: ", err)

		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		log.Println("Can't dial: ", err)

		return nil, err
	}

	client := Client{
		conf:      conf,
		conn:      conn,
		handler:   handler,
		replyByte: make([]byte, protocal.ReplySize),
		headBytes: make([]byte, conf.PacketSize),
	}

	return &client, nil
}

func (c *Client) prepareFile(f string) error {
	var packCount int32

	file, err := os.Open(f)
	if err != nil {
		log.Fatal(err)

		return err
	}

	fileInfo, err := os.Stat(f)
	if err != nil {
		log.Fatal(err)

		return err
	}

	c.file = file
	c.fileInfo = fileInfo

	if fileInfo.Size()%int64(c.conf.PacketSize) != 0 {
		packCount = int32(fileInfo.Size()/int64(c.conf.PacketSize)) + 1
	} else {
		packCount = int32(fileInfo.Size() / int64(c.conf.PacketSize))
	}

	c.proto = &protocal.Proto{
		HeaderSize: int16(protocal.FixedHeaderSize + len(fileInfo.Name())),
		FileSize:   fileInfo.Size(),
		PackSize:   int16(c.conf.PacketSize),
		PackCount:  packCount,
		PackOrder:  0,
	}

	return nil
}

// StartRun - Client start run
func (c *Client) StartRun() (err error) {
	defer func() {
		if err != nil {
			log.Fatal("StartRun - error：", err)
		}
	}()

	err = c.WriteFirst()
	if err != nil {
		return
	}

	_, err = c.conn.Read(c.replyByte)
	if err != nil {
		log.Fatal("StartRun - 文件传输错误：", err)
	}

	if c.replyByte[0] == protocal.ReplyOk {
		for i := int32(1); ; i++ {
			err = c.WriteFile(i)
			if err != nil {
				return
			}

			_, err = c.conn.Read(c.replyByte)
			if err != nil {
				log.Fatal("StartRun - 文件传输错误：", err)
				return
			}

			if c.replyByte[0] == protocal.ReplyOk {
				continue
			}

			log.Println("Server return incorrent.")

			return
		}
	}

	return
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

func (c *Client) convertHeadType() {
	c.headBytes[0] = byte(protocal.HeaderFileType)
}

// WriteFile - 开始向服务器发送文件
func (c *Client) WriteFile(order int32) error {
	protocal.PutInt32(c.headBytes[protocal.PackOrderOffset:], order)
	_, err := c.file.Read(c.headBytes[protocal.FixedHeaderSize:])

	if err != nil {
		if err == io.EOF {
			log.Println("WriteFile - 传送文件结束！")
			c.conn.Close()

			return io.EOF
		}

		log.Fatal("WriteFile - 读取文件错误：", err)

		return err
	}

	_, err = c.conn.Write(c.headBytes)
	if err != nil {
		log.Fatal("WriteFile - 文件传输错误：", err)

		return err
	}

	return nil
}
