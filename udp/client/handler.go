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
	"encoding/binary"
	"hash"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"redalert/protocol"
)

// Handler interface
type Handler interface {
	OnReceive()
	OnProto() error
	OnSend() error
	write() (int, error)
}

// DefaultHandler default handler
type DefaultHandler struct {
	conn  *net.UDPConn
	proto *protocol.Proto

	replyPack []byte
	pack      *protocol.Encode

	hash hash.Hash

	file     *os.File
	fileInfo os.FileInfo

	sendChan chan struct{}
}

// OnProto discuss proto
func (h *DefaultHandler) OnProto() error {
	firstPacket := protocol.Encode{
		Body: make([]byte, protocol.FirstPacketSize),
	}
	firstPacket.Buffer = bytes.NewBuffer(firstPacket.Body)

	firstPacket.Marshal(h.proto)
	h.proto.HeaderType = protocol.HeaderFileType
	h.pack.Marshal(h.proto)
	nameReader := strings.NewReader(h.fileInfo.Name())
	nameReader.Read(firstPacket.Body[protocol.FixedHeaderSize:])

	log.Println("writeFirst - headBytes:", firstPacket.Body[:h.proto.HeaderSize])

	num, err := h.conn.Write(firstPacket.Body)

	if err != nil {
		log.Fatal("writeFirst err:", err)
	}

	log.Println("writeFirst - 写了", num, "个字符。")

	return err
}

// OnReceive receive back bytes
func (h *DefaultHandler) OnReceive() {
	go func() {
		for {
			num, err := h.conn.Read(h.replyPack)
			if err != nil {
				log.Fatal("[ERROR]:readFromUDP - read", num, "byte，", "error：", err)
				return
			}

			packOrder := binary.LittleEndian.Uint32(h.replyPack)

			if protocol.ReplyFinish == packOrder {
				log.Println("Pass file finish.")
				return
			}

			log.Println("[RECEIVE]:Receive length:", num)

			h.proto.PackOrder = packOrder + 1

			log.Println("[RECEIVE]: Pack order:", h.proto.PackOrder)

			h.sendChan <- struct{}{}

			log.Println("[RECEIVE]: Begin send file")
		}
	}()
}

// OnSend send file
func (h *DefaultHandler) OnSend() error {
	binary.BigEndian.PutUint32(h.pack.Body[protocol.PackOrderOffset:], h.proto.PackOrder)
	num, err := h.file.Read(h.pack.Body[protocol.FixedHeaderSize:])

	if err != nil {
		if err == io.EOF {
			log.Println("WriteFile - 传送文件结束！")

			h.hash.Write(h.pack.Body[protocol.FixedHeaderSize : protocol.FixedHeaderSize+num])
			hhash := h.hash.Sum(nil)

			log.Println("文件MD5值：", hhash)

			md5Reader := bytes.NewReader(hhash)
			md5Reader.Read(h.pack.Body[protocol.FixedHeaderSize:])
			h.pack.Body[0] = protocol.HeaderFileFinishType

			h.conn.Write(h.pack.Body)
			h.conn.Close()
			h.file.Close()

			return io.EOF
		}
		log.Fatal("[ERROR]:Read file:", err)

		return err
	}

	log.Println("[SEND]:Read", num, "word.")

	h.hash.Write(h.pack.Body[protocol.FixedHeaderSize : protocol.FixedHeaderSize+num])

	binary.BigEndian.PutUint16(h.pack.Body[protocol.PackSizeOffset:], uint16(num))

	num, err = h.conn.Write(h.pack.Body)
	if err != nil {
		log.Fatal("WriteFile - 写了", num, "个字符。", "文件传输错误：", err)

		return err
	}

	log.Println("WriteFile - 写了", num, "个字符。")

	return nil
}

func (h *DefaultHandler) write() (int, error) {
	return h.conn.Write(h.pack.Body)
}
