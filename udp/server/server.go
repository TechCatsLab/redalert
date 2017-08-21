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

	"redalert/udp/protocol"
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
		udpPort = ":" + protocol.DefaultUDPPort
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

	fmt.Printf("[server] start at %v \n", conn.LocalAddr())

	server := &Service{
		conf:    conf,
		conn:    conn,
		handler: handler,
		buffer:  make([]*Packet, conf.PacketSize),
		sender:  make(chan *Packet, 256),
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
			fmt.Printf("[Send] send a packet to %v \n", pack.Remote)

			if err != nil {
				c.handler.OnError(err, nil)
			}
		}
	}
}

func (c *Service) prepare() {
	c.conn.SetReadBuffer(defaultReadBuffer)
	c.conn.SetWriteBuffer(defaultWriteBuffer)
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
	p := NewPacket(1024)
	reply := make([]byte, 1)

	for {
		size, remote, err := c.conn.ReadFromUDP(p.Body)
		fmt.Printf("[receive] size %d \n", size)
		if err != nil {
			c.handler.OnError(err, remote)
		}
		err = p.Read(c, size, remote)

		if err == nil {
			err = c.handler.OnPacket(p)
			reply[0] = protocol.ReplyOk
			if err == nil && p.Body[0] != protocol.HeaderFileFinishType {
				c.Send(reply, p.Remote)
			}

		}

		if err != nil {
			reply[0] = protocol.ReplyNo
			c.Send(reply, p.Remote)
			c.handler.OnError(err, p.Remote)
		}
	}
}
