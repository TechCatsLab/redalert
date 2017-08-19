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

	"redalert/udp/protocal"
	"redalert/udp/remote"
)

// Packet represent a UDP packet
type Packet struct {
	proto  protocal.Proto
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

func (p *Packet) Read(s *Service, size int, remote *net.UDPAddr) error {
	// size, remote, err := s.conn.ReadFromUDP(p.Body)
	fmt.Printf("[Read] read bytes from %v \n", p.Remote)
	// if err != nil {
	// return err
	// }

	p.Remote = remote
	p.Size = size

	headerType := uint8(p.Body[0])

	switch {
	case headerType == protocal.HeaderRequestType:
		return p.handleRequest(s)

	case headerType == protocal.HeaderFileType:
		return p.handleFilePacket(s)

	case headerType == protocal.HeaderFileFinishType:
		return p.handleFileFinishPacket(s, size)
	}

	return nil
}

// Reset the buffer
func (p *Packet) Reset() {
	p.Size = 0
	p.Remote = nil
}

func (p *Packet) handleRequest(s *Service) error {
	headerSize := binary.LittleEndian.Uint16(p.Body[protocal.HeaderSizeOffset:protocal.FileSizeOffset])
	fmt.Printf("[handleRequest] header size is %d \n", headerSize)
	filename := string(p.Body[protocal.FileNameOffset:headerSize])
	fmt.Printf("file name is %s \n", filename)
	if rem, ok := remote.Service.GetRemote(p.Remote); ok {
		if filename != rem.FileName {
			return ErrDiffrentFile
		}

		return ErrDuplicated
	}

	file, err := os.Create(protocal.DefaultDir + filename)
	if err != nil {
		return err
	}

	if ok := remote.Service.OnStartTransfor(filename, file, p.Remote); !ok {
		return ErrDiffrentFile
	}

	p.proto.HeaderType = uint8(p.Body[0])

	return nil
}

func (p *Packet) handleFilePacket(s *Service) error {
	fmt.Printf("[handleFilePacket] receive pack file from %v \n", p.Remote)
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		return ErrInvalidFilePack
	}

	p.proto.HeaderType = uint8(p.Body[0])

	order := binary.LittleEndian.Uint32(p.Body[protocal.PackOrderOffset:protocal.FixedHeaderSize])
	if order-rem.PackCount > 1 {
		return ErrInvalidOrder
	}

	realBody := p.Body[protocal.FixedHeaderSize:]

	n, err := rem.File.Write(realBody)
	if err != nil {
		return err
	}

	if n < len(realBody) {
		return ErrWrite
	}

	return nil
}

func (p *Packet) handleFileFinishPacket(s *Service, n int) error {
	fmt.Printf("[handleFileFinishPacket] client %v finish sending \n", p.Remote)
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		fmt.Printf("query from map-> file name is %v \n", rem)
		return ErrNotExists
	}

	p.proto.HeaderType = uint8(p.Body[0])

	h := string(p.Body[1:n])
	hash := rem.Hash.Sum(nil)

	fmt.Printf("jsharkc hash is %v and my hash is %v \n", p.Body[protocal.FixedHeaderSize:protocal.FixedHeaderSize+32], hash)

	if string(hash) != h {
		return ErrHashNotMatch
	}

	return nil
}
