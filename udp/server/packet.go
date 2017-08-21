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
*     Initial: 2017/08/15          Yusan Kurban
 */

package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"

	"redalert/udp/protocol"
	"redalert/udp/remote"
)

// Packet represent a UDP packet
type Packet struct {
	proto  protocol.Proto
	Body   []byte
	Size   int
	Remote *net.UDPAddr
}

var (
	ErrDuplicated      = errors.New("Duplicated request for one file")
	ErrDiffrentFile    = errors.New("Client try to send multi file")
	ErrInvalidFilePack = errors.New("Invalid file packet")
	ErrInvalidOrder    = errors.New("Invalid pack order")
	ErrWrite           = errors.New("write file fail")
	ErrReset           = errors.New("failed with reset timer")
	ErrNotExists       = errors.New("Remote Address not exists")
	ErrHashNotMatch    = errors.New("hash value not match")
)

// NewPacket generates a Packet with a len([]byte) == cap.
func NewPacket(cap int) *Packet {
	return &Packet{
		Body:   make([]byte, cap),
		Size:   0,
		Remote: nil,
	}
}

// WriteToUDP write packet to UDP
func (p *Packet) WriteToUDP(conn *net.UDPConn) error {
	_, err := conn.WriteToUDP(p.Body[:p.Size], p.Remote)

	fmt.Printf("[WriteToUDP] send %v to %v \n", p.Body, *p.Remote)

	return err
}

// Read read packet and handle on the base of type
// TODO: 考虑把从 UDP 中读取操作从 receive() 函数放回到 Read() 函数中
func (p *Packet) Read(s *Service, size int, remote *net.UDPAddr) error {
	var err error
	// size, remote, err := s.conn.ReadFromUDP(p.Body)
	fmt.Printf("[Read] read bytes from %v \n", p.Remote)
	// if err != nil {
	// return err
	// }

	p.Remote = remote
	p.Size = size

	headerType := uint8(p.Body[0])

	switch {
	case headerType == protocol.HeaderRequestType:
		err = p.handleRequest(s)

	case headerType == protocol.HeaderFileType:
		err = p.handleFilePacket(s)

	case headerType == protocol.HeaderFileFinishType:
		err = p.handleFileFinishPacket(s)
	}

	if err != nil {
		return err
	}

	p.proto.HeaderType = uint8(p.Body[0])

	return nil
}

// Reset the buffer
func (p *Packet) Reset() {
	p.Size = 0
	p.Remote = nil
}

func (p *Packet) handleRequest(s *Service) error {
	headerSize := binary.LittleEndian.Uint16(p.Body[protocol.HeaderSizeOffset:protocol.FileSizeOffset])
	filename := string(p.Body[protocol.FileNameOffset:headerSize])

	fmt.Printf("[handleRequest] header size is %d \n", headerSize)
	fmt.Printf("file name is %s \n", filename)

	if rem, ok := remote.Service.GetRemote(p.Remote); ok {
		if filename != rem.FileName {
			return ErrDiffrentFile
		}

		return ErrDuplicated
	}

	file, err := os.Create(protocol.DefaultDir + filename)
	if err != nil {
		return err
	}

	if ok := remote.Service.OnStartTransfor(filename, file, p.Remote); !ok {
		return ErrDiffrentFile
	}

	return nil
}

func (p *Packet) handleFilePacket(s *Service) error {
	order := binary.LittleEndian.Uint32(p.Body[protocol.PackOrderOffset:protocol.FixedHeaderSize])
	availableSize := binary.LittleEndian.Uint16(p.Body[protocol.PackSizeOffset:protocol.PackCountOffset])
	realBody := p.Body[protocol.FixedHeaderSize : protocol.FixedHeaderSize+availableSize]

	fmt.Printf("[handleFilePacket] receive pack file from %v \n", p.Remote)

	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		return ErrInvalidFilePack
	}

	fmt.Printf("[ORDER] is %d \n", order)

	if order == rem.PackCount {
		return nil
	}

	if order-rem.PackCount > 1 {
		return ErrInvalidOrder
	}

	n, err := rem.File.Write(realBody)
	if err != nil {
		return err
	}

	if n < len(realBody) {
		return ErrWrite
	}

	p.proto.PackSize = availableSize

	return nil
}

func (p *Packet) handleFileFinishPacket(s *Service) error {
	fmt.Printf("[handleFileFinishPacket] client %v finish sending \n", p.Remote)
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		fmt.Printf("query from map-> file name is %v \n", rem)
		return ErrNotExists
	}

	hash := rem.Hash.Sum(nil)

	fmt.Print("receive finish \n")

	if string(p.Body[protocol.FixedHeaderSize:protocol.FixedHeaderSize+16]) != string(hash) {
		return ErrHashNotMatch
	}

	remote.Service.Close(p.Remote, nil)

	return nil
}
