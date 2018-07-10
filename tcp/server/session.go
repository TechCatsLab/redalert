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
	"encoding/binary"
	"bytes"
	"hash"
	"log"
	"net"
	"os"

	"github.com/TechCatsLab/redalert/protocol"
)

// Session a connection
type Session struct {
	Pack      *protocol.Encode
	Reply     []byte
	file      *os.File
	conn      net.Conn
	proto     *protocol.Proto
	CountChan chan bool
	hash      hash.Hash
}

// Start start write file
func (s *Session) Start() {
	packOrder := uint32(1)
	for {
		num, err := s.conn.Read(s.Pack.Body)
		if err != nil {
			log.Println("[ERROR]:Read connect error", err)

			s.file.Close()
			os.Remove(s.file.Name())
			s.CountChan <- false
			return
		}

		log.Printf("[DEBUG]:Read %d bytes.", num)

		s.Pack.Buffer = bytes.NewBuffer(s.Pack.Body)
		s.Pack.Unmarshal(s.proto)

		if s.Pack.Body[0] == protocol.HeaderFileFinishType {
			md5hash := s.hash.Sum(nil)
			//if string(md5hash) != string(s.Pack.Body[protocol.FixedHeaderSize:protocol.FixedHeaderSize+16]) {
			//	log.Println("[DEBUG]:MD5 error.")
			//
			//	s.file.Close()
			//	os.Remove(s.file.Name())
			//}

			log.Printf("[DEBUG]:Recive file finish.hash %x", md5hash)

			s.file.Close()
			s.CountChan <- false
			return
		}

		log.Println("[DEBUG]:Before judge order", s.proto.PackOrder)

		if s.proto.PackOrder != packOrder {
			binary.BigEndian.PutUint32(s.Reply, packOrder)
			_, err := s.conn.Write(s.Reply)
			if err != nil {
				log.Println("[ERROR]:Conn write error", err)

				s.file.Close()
				os.Remove(s.file.Name())
				s.CountChan <- false

				return
			}

			log.Println("[DEBUG]:Resend order", s.proto.PackOrder)
			continue
		}

		log.Printf("[DEBUG]:PackSize %d, Pack length %d", s.proto.PackSize, len(s.Pack.Body))
		realBody := s.Pack.Body[protocol.FixedHeaderSize:num]

		s.file.Write(realBody)
		s.hash.Write(realBody)

		binary.BigEndian.PutUint32(s.Reply, packOrder)
		_, err = s.conn.Write(s.Reply)
		if err != nil {
			log.Println("[ERROR]:Conn write error", err)

			s.file.Close()
			os.Remove(s.file.Name())
			s.CountChan <- false

			return
		}

		log.Printf("[DEBUG]:Handler %d pack.", packOrder)
		packOrder++
	}
}
