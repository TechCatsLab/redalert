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
	"crypto/md5"
	"encoding/binary"
	"log"
	"net"
	"os"
	"redalert/protocol"
	"time"
)

// Server tcp server
type Server struct {
	conf      *Conf
	totalConn int
	CountChan chan bool
}

// StartServer start a new TCP server
func StartServer(conf *Conf) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", conf.Addr+":"+conf.Port)
	if err != nil {
		log.Fatalln("[ERROR]:Can't resolve address:", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalln("[ERROR]:Can't resolve address:", err)
	}

	s := &Server{
		conf:      conf,
		totalConn: 0,
		CountChan: make(chan bool),
	}

	countConn(s)

	for {
		if s.totalConn >= s.conf.MaxConn {
			time.Sleep(time.Second * 2)
		} else {
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.Println("[ERROR]:listen error", err)
			} else {
				s.onConn(conn)
			}
		}
	}
}

func (s *Server) onConn(conn *net.TCPConn) {
	pack := make([]byte, protocol.FirstPacketSize)

	proto := protocol.Proto{
		HeaderType: pack[0],
		HeaderSize: binary.LittleEndian.Uint16(pack[protocol.HeaderSizeOffset:protocol.FileSizeOffset]),
		PackSize:   binary.LittleEndian.Uint16(pack[protocol.PackSizeOffset:protocol.PackCountOffset]),
		PackCount:  binary.LittleEndian.Uint32(pack[protocol.PackCountOffset:protocol.PackOrderOffset]),
		PackOrder:  binary.LittleEndian.Uint32(pack[protocol.PackOrderOffset:protocol.FixedHeaderSize]),
	}

	log.Println("[CONN]:Begin create file")

	filename := string(pack[protocol.FileNameOffset:proto.HeaderSize])

	file, err := os.Create(protocol.DefaultDir + filename)
	if err != nil {
		log.Println("[ERROR]:Create file error", err)
		return
	}

	session := Session{
		Pack:      make([]byte, proto.PackSize),
		Reply:     make([]byte, protocol.ReplySize),
		file:      file,
		proto:     &proto,
		CountChan: s.CountChan,
		hash:      md5.New(),
	}

	s.CountChan <- true
	session.Start()
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

			if s.totalConn == 0 {
				os.Exit(0)
			}
		}
	}()
}
