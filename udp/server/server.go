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

package server

import (
	"fmt"
	"net"

	"redalert/udp/protocal"
)

const (
	defaultReadBuffer  = 65536
	defaultWriteBuffer = 65536
)

// Conf represents the UDP server configuration, such as IP, port, etc.
type Conf struct {
	Address    string // Local Address
	Port       string // Local Port
	PacketSize int    // Packet max size
	CacheCount int    // Cache size
}

// Service is a UDP service
type Service struct {
	conf    *Conf
	conn    *net.UDPConn
	handler Handler
	buffer  []*Packet
	sender  chan *Packet
	close   chan struct{}
}

// NewServer start a new UDP service
func NewServer(conf *Conf, handler Handler) (*Service, error) {
	var udpPort string

	if conf.Port == "" {
		udpPort = ":" + protocal.DefaultUDPPort
	} else {
		udpPort = conf.Address + ":" + conf.Port
	}

	addr, err := net.ResolveUDPAddr("udp", udpPort)
	if err != nil {
		fmt.Println("Can't resolve addr:", err.Error())
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("listen error:", err.Error())
		return nil, err
	}

	server := &Service{
		conf:    conf,
		conn:    conn,
		handler: handler,
		buffer:  make([]*Packet, conf.PacketSize),
		sender:  make(chan *Packet),
		close:   make(chan struct{}),
	}
	server.prepare()

	go server.handlerClient()

	return server, nil
}

func (c *Service) handlerClient() {
	go c.receive()

	for {
		select {
		case <-c.close:
			c.conn.Close()

		case pack := <-c.sender:
			err := pack.WriteToUDP(c.conn)

			if err != nil {
				c.handler.OnError(err, nil)
			}
		}
	}
}

func (c *Service) prepare() {
	c.conn.SetReadBuffer(defaultReadBuffer)
	c.conn.SetWriteBuffer(defaultWriteBuffer)

	for i := 0; i < c.conf.CacheCount; i++ {
		c.buffer[i] = NewPacket(c.conf.PacketSize)
	}
}

// Close close current service
func (c *Service) Close() {
	c.close <- struct{}{}
}

// Send send a packet to remote
func (c *Service) Send(body []byte, remote *net.UDPAddr) {
	pack := &Packet{
		Body:   body,
		Size:   len(body),
		Remote: remote,
	}

	c.sender <- pack
}

func (c *Service) receive() {
	// index := 0
	// cap := c.conf.CacheCount

	for {
		// Pick a packet and reset it for receiving.
		// packet := c.buffer[index]
		// packet.Reset()
		p := NewPacket(1024)
		size, remote, err := c.conn.ReadFromUDP(p.Body)
		if err != nil {
			c.handler.OnError(err, remote)
		}
		err = p.Read(c, size, remote)
		reply := make([]byte, 1)

		if err != nil {
			reply[0] = protocal.ReplyNo
			c.Send(reply, p.Remote)
			c.handler.OnError(err, p.Remote)
		} else {
			reply[0] = protocal.ReplyOk
			c.Send(reply, p.Remote)
			c.handler.OnPacket(p)
		}

		// index = (index + 1) % cap
	}
}
