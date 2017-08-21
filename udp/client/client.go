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
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"redalert/udp/protocol"
)

const resendInterval = 200

var (
	// ErrLittleHead little header
	ErrLittleHead = errors.New("Bytes of header too little to write")
)

// Client - UDP Client
type Client struct {
	conf  *Conf
	conn  *net.UDPConn
	proto *protocol.Proto

	replyPack []byte
	headPack  []byte

	hash hash.Hash

	file     *os.File
	fileInfo os.FileInfo

	sendChan chan struct{}
	receChan chan uint32
}

// NewClient 创建一个 UDP 客户端
func NewClient(conf *Conf) (*Client, error) {
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
		hash:      md5.New(),
		replyPack: make([]byte, protocol.ReplySize),
		headPack:  make([]byte, conf.PacketSize),
		sendChan:  make(chan struct{}, 1),
		receChan:  make(chan uint32),
	}

	return &client, nil
}

// PrepareFile - prepare file.
func (c *Client) PrepareFile(f string) error {
	var packCount uint32

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

	if uint64(fileInfo.Size())%uint64(c.conf.PacketSize) != 0 {
		packCount = uint32(fileInfo.Size()/int64(c.conf.PacketSize)) + 1
	} else {
		packCount = uint32(fileInfo.Size() / int64(c.conf.PacketSize))
	}

	c.proto = &protocol.Proto{
		HeaderSize: uint16(protocol.FixedHeaderSize + len(fileInfo.Name())),
		FileSize:   uint64(fileInfo.Size()),
		PackSize:   uint16(c.conf.PacketSize),
		PackCount:  packCount,
		PackOrder:  0,
	}

	return nil
}

// Start - Client start run
func (c *Client) Start() (err error) {
	defer func() {
		if err != nil {
			log.Fatal("Start - error：", err)
		}
	}()

	go readFromUDP(c)

	begin := time.Now()

	err = c.writeFirst()
	if err != nil {
		return
	}

	for {
		select {
		case packOrder := <-c.receChan:
			if packOrder == c.proto.PackOrder {
				c.proto.PackOrder++
			} else {
				c.proto.PackOrder = packOrder + 1
			}

			c.sendChan <- struct{}{}
		case <-c.sendChan:
			err := c.writeFile()

			if err != nil {
				if err == io.EOF {
					err = nil
					log.Println("Time:", time.Now().Sub(begin))

					return err
				}

				log.Fatal("Start - Error：", err)
			}
		case <-time.After(resendInterval * time.Millisecond):
			num, err := c.conn.Write(c.headPack)
			if err != nil {
				log.Fatal("WriteFile - 写了", num, "个字符。", "文件传输错误：", err)

				return err
			}

			log.Println("WriteFile - 写了", num, "个字符。")

			return nil
		}
	}
}

func readFromUDP(c *Client) {
	for {
		num, err := c.conn.Read(c.replyPack)
		if err != nil {
			log.Fatal("readFromUDP - read", num, "byte，", "error：", err)
			return
		}
		log.Println("num:", num)
		packOrder := binary.LittleEndian.Uint32(c.replyPack)

		if protocol.ReplyFinish == packOrder {
			log.Println("Pass file finish.")
			return
		}

		c.receChan <- packOrder
	}
}

func (c *Client) writeHead(b []byte) error {
	if len(b) < protocol.FixedHeaderSize {
		return ErrLittleHead
	}

	binary.LittleEndian.PutUint16(b[protocol.HeaderSizeOffset:], c.proto.HeaderSize)
	binary.LittleEndian.PutUint64(b[protocol.FileSizeOffset:], c.proto.FileSize)
	binary.LittleEndian.PutUint16(b[protocol.PackSizeOffset:], c.proto.PackSize)
	binary.LittleEndian.PutUint32(b[protocol.PackCountOffset:], c.proto.PackCount)
	binary.LittleEndian.PutUint32(b[protocol.PackOrderOffset:], c.proto.PackOrder)

	return nil
}

// WriteFirst - 第一次连接服务器，商量协议
func (c *Client) writeFirst() error {
	firstPacket := make([]byte, protocol.FirstPacketSize)

	err := c.writeHead(firstPacket)
	c.writeHead(c.headPack)
	firstPacket[0] = byte(protocol.HeaderRequestType)
	c.headPack[0] = byte(protocol.HeaderFileType)
	nameReader := strings.NewReader(c.fileInfo.Name())
	nameReader.Read(firstPacket[protocol.FixedHeaderSize:])

	log.Println("writeFirst - headBytes:", firstPacket[:c.proto.HeaderSize])

	num, err := c.conn.Write(firstPacket)

	if err != nil {
		log.Fatal("writeFirst err:", err)
	}

	log.Println("writeFirst - 写了", num, "个字符。")

	return err
}

// writeFile - 开始向服务器发送文件
func (c *Client) writeFile() error {
	binary.LittleEndian.PutUint32(c.headPack[protocol.PackOrderOffset:], c.proto.PackOrder)
	num, err := c.file.Read(c.headPack[protocol.FixedHeaderSize:])

	if err != nil {
		log.Fatal("WriteFile - 读到", num, "个字符。", "读取文件错误：", err)

		return err
	}

	if num < c.conf.PacketSize-protocol.FixedHeaderSize {
		log.Println("WriteFile - 传送文件结束！")

		c.hash.Write(c.headPack[protocol.FixedHeaderSize : protocol.FixedHeaderSize+num])
		h := c.hash.Sum(nil)

		log.Println("文件MD5值：", h)

		md5Reader := bytes.NewReader(h)
		md5Reader.Read(c.headPack[protocol.FixedHeaderSize:])
		c.headPack[0] = protocol.HeaderFileFinishType

		c.conn.Write(c.headPack)
		c.conn.Close()
		c.file.Close()
		close(c.receChan)

		return io.EOF
	}

	log.Println("WriteFile - 读到", num, "个字符。")

	c.hash.Write(c.headPack[protocol.FixedHeaderSize:])

	binary.LittleEndian.PutUint16(c.headPack[protocol.PackSizeOffset:], uint16(num))

	num, err = c.conn.Write(c.headPack)
	if err != nil {
		log.Fatal("WriteFile - 写了", num, "个字符。", "文件传输错误：", err)

		return err
	}

	log.Println("WriteFile - 写了", num, "个字符。")

	return nil
}
