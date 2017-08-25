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
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

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
		info:   &FileInfo{},
	}

	client.prepareBuffer()

	if err = client.initFile(conf.FileName); err != nil {
		fmt.Printf("[ERROR] initFile crash with error: %v \n", err)
	}

	return client, nil
}
func (c *Client) prepareBuffer() {
	c.conn.SetReadBuffer(bufferSize)
	c.conn.SetWriteBuffer(bufferSize)
}

func (c *Client) initFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	c.info.file = file
	c.info.fileName = fileInfo
	c.info.hash = md5.New()

	return nil
}

func (c *Client) consult() error {
	c.proto.FileSize = uint64(c.info.fileName.Size())
	c.proto.HeaderSize = protocol.HeaderSize
	c.proto.PackSize = protocol.FirstPacketSize
	c.info.headPack = make([]byte, protocol.FirstPacketSize)

	err := c.packHead(c.info.headPack)
	if err != nil {
		return err
	}

	c.info.headPack[0] = byte(protocol.HeaderRequestType)
	nameReader := strings.NewReader(c.info.fileName.Name())
	nameReader.Read(c.info.headPack[protocol.FixedHeaderSize:])

	n, err := c.conn.Write(c.info.headPack)
	if err != nil {
		fmt.Printf("[ERROR] Write consult pack error:%v \n", err)
		return err
	}

	fmt.Printf("[WRITE] write %d byte \n", n)

	return nil
}

func (c *Client) send(pack []byte) error {
	n, err := c.conn.Write(pack)
	if err != nil {
		return err
	}

	if n < len(pack) {
		return errWriteIncomplete
	}

	return nil
}

func (c *Client) packHead(b []byte) error {
	if len(b) < protocol.FixedHeaderSize {
		return errInvalidHeaderSize
	}

	binary.LittleEndian.PutUint16(b[protocol.HeaderSizeOffset:], c.proto.HeaderSize)
	binary.LittleEndian.PutUint64(b[protocol.FileSizeOffset:], c.proto.FileSize)
	binary.LittleEndian.PutUint16(b[protocol.PackSizeOffset:], c.proto.PackSize)
	binary.LittleEndian.PutUint32(b[protocol.PackCountOffset:], c.proto.PackCount)
	binary.LittleEndian.PutUint32(b[protocol.PackOrderOffset:], c.proto.PackOrder)

	return nil
}

func (c *Client) receive(conn *net.TCPConn) {
	go func() {
		num, err := conn.Read(c.info.replyPack)
		if err != nil {
			fmt.Printf("[ERROR]:readFromUDP - read num %d byte error：%v \n", num, err)
			return
		}

		packOrder := binary.LittleEndian.Uint32(c.info.replyPack)

		if protocol.ReplyFinish == packOrder {
			log.Println("Pass file finish.")
			return
		}

		log.Println("[RECEIVE]:Receive length:", num)

		c.proto.PackOrder = packOrder + 1

		fmt.Printf("[RECEIVE]: Pack order: %d \n", c.proto.PackOrder)

	}()
}
