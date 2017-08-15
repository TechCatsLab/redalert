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

// Packet represent a UDP packet
type Packet struct {
	Body   []byte
	Size   int
	Remote *net.UDPAddr
}

// Service is a UDP service
type Service struct {
	Conn    *net.UDPConn
	handler Handler
	sender  chan *Packet
	close   chan struct{}
}

// NewService start a new UDP service
func NewService(port string, handler Handler) (*Service, error) {
	var udpPort string

	if port == "" {
		udpPort = ":" + protocal.DefaultUDPPort
	} else {
		udpPort = ":" + port
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
		Conn:    conn,
		handler: handler,
		sender:  make(chan *Packet),
		close:   make(chan struct{}),
	}

	go server.handlerClient()

	return server, nil
}

func (c *Service) handlerClient() {
	for {
		select {
		case <-c.close:
			c.Conn.Close()

		case pack := <-c.sender:
			err := pack.WriteToUDP(c.Conn)

			if err != nil {
				c.handler.OnError(err)
			}
		}
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

// WriteToUDP write packet to UDP
func (p *Packet) WriteToUDP(conn *net.UDPConn) error {
	_, err := conn.WriteToUDP(p.Body[:p.Size], p.Remote)

	return err
}
