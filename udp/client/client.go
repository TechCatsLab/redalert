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
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"redalert/protocol"
)

const (
	resendInterval = 500
	bufferSize     = 65535
)

var (
	// ErrLittleHead little header
	ErrLittleHead = errors.New("Bytes of header too little to write")
)

// Client - UDP Client
type Client struct {
	conf    *Conf
	proto   *protocol.Proto
	handler Handler

	sendChan chan struct{}
}

// NewClient Create a UDP client
func NewClient(conf *Conf) (*Client, error) {
	addr, err := net.ResolveUDPAddr("udp", conf.RemoteAddress+":"+conf.RemotePort)
	if err != nil {
		log.Fatalln("[ERROR]:Can't resolve address:", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		log.Fatalln("[ERROR]:Can't dial:", err)
	}

	conn.SetReadBuffer(bufferSize)
	conn.SetWriteBuffer(bufferSize)

	file, err := os.Open(conf.FileName)
	if err != nil {
		log.Fatalln("[ERROR]:Open file:", err)
	}

	fileInfo, err := os.Stat(conf.FileName)
	if err != nil {
		log.Fatalln("[ERROR]:Get fileinfo:", err)
	}

	if conf.PacketSize < protocol.FirstPacketSize {
		conf.PacketSize = protocol.FirstPacketSize
	}

	if conf.PacketSize > protocol.MaxPacketSize {
		conf.PacketSize = protocol.MaxPacketSize
	}

	client := Client{
		conf:     conf,
		sendChan: make(chan struct{}),
	}

	client.proto = &protocol.Proto{
		HeaderType: protocol.HeaderRequestType,
		HeaderSize: uint16(protocol.FixedHeaderSize + len(fileInfo.Name())),
		PackSize:   uint16(client.conf.PacketSize),
		PackOrder:  0,
	}

	decode := protocol.Encode{
		Body: make([]byte, conf.PacketSize),
	}

	decode.Buffer = bytes.NewBuffer(decode.Body)

	handler := &DefaultHandler{
		conn:      conn,
		proto:     client.proto,
		hash:      md5.New(),
		replyPack: make([]byte, protocol.ReplySize),
		pack:      &decode,
		sendChan:  client.sendChan,
		file:      file,
		fileInfo:  fileInfo,
	}

	client.handler = handler

	return &client, nil
}

// Start - Client start run
func (c *Client) Start() (err error) {
	defer func() {
		if err != nil {
			log.Fatal("[ERROR]:Start", err)
		}
	}()

	c.handler.OnReceive()

	begin := time.Now()

	err = c.handler.OnProto()
	if err != nil {
		return
	}

	for {
		select {
		case <-c.sendChan:
			err := c.handler.OnSend()

			if err != nil {
				if err == io.EOF {
					err = nil
					log.Println("[TIME]:", time.Now().Sub(begin))

					return err
				}

				log.Fatal("[ERROR]: Send error", err)
			}
		case <-time.After(resendInterval * time.Millisecond):
			num, err := c.handler.write()
			if err != nil {
				log.Fatal("[ERROR]: Resend", num, "word.", err)

				return err
			}

			log.Println("[RESEND]: ", num, "word.")
		}
	}
}
