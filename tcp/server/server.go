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
 *     Initial: 2017/08/26        Liu JiaChang
 */

package server

import (
	"bytes"
	"crypto/md5"
	"log"
	"net"
	"os"
	"time"

	"github.com/TechCatsLab/redalert/protocol"
)

// Server tcp server
type Server struct {
	conf      *Conf
	totalConn int
	CountChan chan bool
	listener  *net.TCPListener
}

// NewServer start a new TCP server
func NewServer(conf *Conf) *Server {
	tcpAddr, err := net.ResolveTCPAddr("tcp", conf.Addr+":"+conf.Port)
	if err != nil {
		log.Fatalln("[ERROR]:Can't resolve address:", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalln("[ERROR]:Can't resolve address:", err)
	}

	log.Printf("[server] start at %s", tcpAddr.String())

	s := &Server{
		conf:      conf,
		totalConn: 0,
		CountChan: make(chan bool),
		listener:  listener,
	}

	return s
}

// Start TCP server
func (s *Server) Start() {
	countConn(s)

	for {
		if s.totalConn >= s.conf.MaxConn {
			time.Sleep(time.Second * 2)
		} else {
			conn, err := s.listener.AcceptTCP()
			if err != nil {
				log.Println("[ERROR]:listen error", err)
			} else {
				s.onConn(conn)
			}
		}
	}
}

func (s *Server) onConn(conn *net.TCPConn) {
	firstDecode := protocol.Encode{
		Body: make([]byte, protocol.FirstPacketSize),
	}

	firstDecode.Buffer = bytes.NewBuffer(firstDecode.Body)

	_, err := conn.Read(firstDecode.Body)
	if err != nil {
		log.Println("[ERROR]:Conn read error", err)
		return
	}

	proto := protocol.Proto{}

	firstDecode.Unmarshal(&proto)

	log.Printf("[CONN]:Begin create file, Proto: %#v\n", proto)

	filename := string(firstDecode.Body[protocol.FileNameOffset:proto.HeaderSize])

	file, err := os.Create(protocol.DefaultDir + filename)
	if err != nil {
		log.Println("[ERROR]:Create file error", err)
		return
	}

	log.Printf("[DEBUG]:File name %s", filename)

	decode := protocol.Encode{
		Body: make([]byte, proto.PackSize),
	}

	decode.Buffer = bytes.NewBuffer(decode.Body)

	session := Session{
		Pack:      &decode,
		Reply:     make([]byte, protocol.ReplySize),
		file:      file,
		conn:      conn,
		proto:     &proto,
		CountChan: s.CountChan,
		hash:      md5.New(),
	}

	num, err := conn.Write(session.Reply)
	if err != nil {
		log.Printf("[ERROR]:Conn write %d word, error %v", num, err)

		return
	}

	s.CountChan <- true
	go session.Start()
}

func countConn(s *Server) {
	go func() {
		for {
			count := <-s.CountChan
			if count {
				s.totalConn++
			} else {
				s.totalConn--
			}

			//if s.totalConn == 0 {
			//	os.Exit(0)
			//}
		}
	}()
}
