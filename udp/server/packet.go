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

	"redalert/protocol"
	"redalert/udp/remote"
)

// Packet represent a UDP packet
type Packet struct {
	proto  protocol.Proto
	Body   []byte
	Size   int
	Repeat uint8 // flag of packet is if repeat packet
	Remote *net.UDPAddr
}

var (
	//ErrDuplicated error of duplicated request
	ErrDuplicated = errors.New("Duplicated request")
	// ErrInvalidFilePack error for client try send file pack before request
	ErrInvalidFilePack = errors.New("Invalid file packet")
	// ErrInvalidOrder error for when pack order is mess
	ErrInvalidOrder = errors.New("Invalid pack order")
	// ErrWrite error for write file error
	ErrWrite = errors.New("write file fail")
	// ErrReset error for reset timer failed
	ErrReset = errors.New("failed with reset timer")
	// ErrNotExists for client cannot find client address from client table
	ErrNotExists = errors.New("Remote Address not exists")
	// ErrHashNotMatch error for hash from client not match with hash which calculated by server
	ErrHashNotMatch = errors.New("hash value not match")
)

// NewPacket generates a Packet with a len([]byte) == cap.
func NewPacket(cap int) *Packet {
	return &Packet{
		Body:   make([]byte, cap),
		Size:   0,
		Remote: nil,
		proto:  protocol.Proto{},
	}
}

// WriteToUDP write packet to UDP
func (p *Packet) WriteToUDP(conn *net.UDPConn) error {
	_, err := conn.WriteToUDP(p.Body[:p.Size], p.Remote)

	fmt.Printf("[WriteToUDP] send %v to %v \n", p.Body, *p.Remote)

	return err
}

// Read read packet and handle on the base of type
func (p *Packet) Read(s *Service, size int, remote *net.UDPAddr) error {
	var err error
	// size, remote, err := s.conn.ReadFromUDP(p.Body)
	fmt.Printf("[Read] read bytes from %v \n", remote)
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

// resolve request type pack and add the client who send this pack to online table
func (p *Packet) handleRequest(s *Service) error {
	headerSize := binary.LittleEndian.Uint16(p.Body[protocol.HeaderSizeOffset:protocol.FileSizeOffset])
	packSize := binary.LittleEndian.Uint16(p.Body[protocol.PackSizeOffset:protocol.PackCountOffset])
	filename := string(p.Body[protocol.FileNameOffset:headerSize])

	fmt.Printf("[handleRequest] header size is %d and packet size is %d \n", headerSize, packSize)
	fmt.Printf("file name is %s \n", filename)

	if _, ok := remote.Service.GetRemote(p.Remote); ok {
		return ErrDuplicated
	}

	file, err := os.Create(protocol.DefaultDir + filename)
	if err != nil {
		return err
	}

	p.Body = make([]byte, packSize)
	p.Body[0] = protocol.HeaderRequestType
	remote.Service.OnStartTransfor(filename, file, p.Remote)

	return nil
}

// resolve file type pack and write the content of pack to file
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
	fmt.Printf("[ORDER] in map is %d \n", rem.PackCount)

	if order == rem.PackCount {
		fmt.Printf("[Repeat packet] %d \n", order)
		p.proto.PackOrder = order
		p.Repeat = 1
		return nil
	}

	p.Repeat = 0

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
	p.proto.PackOrder = order

	return nil
}

// when file transfer finish, calculate hash of file and compare with hash send by client
func (p *Packet) handleFileFinishPacket(s *Service) error {
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

	p = pack
	p.proto.PackOrder = 0
	p.proto.PackSize = 0

	return nil
}
