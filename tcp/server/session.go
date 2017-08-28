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
	"hash"
	"log"
	"net"
	"os"
	"redalert/protocol"
)

// Session a connection
type Session struct {
	Pack      []byte
	Reply     []byte
	file      *os.File
	conn      net.Conn
	proto     *protocol.Proto
	CountChan chan bool
	hash      hash.Hash
}

// Start start write file
func (s *Session) Start() {
	go func() {
		packOrder := uint32(1)
		for {
			_, err := s.conn.Read(s.Pack)
			if err != nil {
				log.Println("[ERROR]:Read connect error", err)

				s.file.Close()
				os.Remove(s.file.Name())
				s.CountChan <- false
				return
			}

			if s.Pack[0] == protocol.HeaderFileFinishType {
				md5hash := s.hash.Sum(nil)
				if string(md5hash) != string(s.Pack[protocol.FixedHeaderSize:protocol.FixedHeaderSize+16]) {
					s.file.Close()
					os.Remove(s.file.Name())
				}

				s.file.Close()
				s.CountChan <- false
				return
			}

			order := binary.LittleEndian.Uint32(s.Pack[protocol.PackOrderOffset:protocol.FixedHeaderSize])

			if order != packOrder {
				binary.LittleEndian.PutUint32(s.Reply, packOrder)
				_, err := s.conn.Write(s.Reply)
				if err != nil {
					log.Println("[ERROR]:Conn write error", err)

					s.file.Close()
					os.Remove(s.file.Name())
					s.CountChan <- false

					return
				}

				continue
			}

			availableSize := binary.LittleEndian.Uint16(s.Pack[protocol.PackSizeOffset:protocol.PackCountOffset])
			realBody := s.Pack[protocol.FixedHeaderSize : protocol.FixedHeaderSize+availableSize]

			s.file.Write(realBody)
			s.hash.Write(realBody)

			packOrder++
		}
	}()
}
