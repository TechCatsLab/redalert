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
	"errors"
	"fmt"
	"net"
	"os"

	"bytes"
	"redalert/protocol"
	"redalert/udp/remote"
)

// Packet represent a UDP packet
type Packet struct {
	proto  *protocol.Proto
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
		proto:  &protocol.Proto{},
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

	requestPack := protocol.Encode{
		Body: p.Body,
	}
	requestPack.Buffer = bytes.NewBuffer(requestPack.Body)

	// unmarshal to p.proto
	requestPack.Unmarshal(p.proto)

	fmt.Println(requestPack)

	p.Remote = remote
	p.Size = size

	switch {
	case p.proto.HeaderType == protocol.HeaderRequestType:
		err = p.handleRequest(s)

	case p.proto.HeaderType == protocol.HeaderFileType:
		err = p.handleFilePacket(s)

	case p.proto.HeaderType == protocol.HeaderFileFinishType:
		err = p.handleFileFinishPacket(s)
	}

	return err
}

// resolve request type pack and add the client who send this pack to online table
func (p *Packet) handleRequest(s *Service) error {
	filename := string(p.Body[protocol.FileNameOffset:p.proto.HeaderSize])

	if _, ok := remote.Service.GetRemote(p.Remote); ok {
		return ErrDuplicated
	}

	file, err := os.Create(protocol.DefaultDir + filename)
	if err != nil {
		return err
	}

	p.Body = make([]byte, p.proto.PackSize)
	remote.Service.OnStartTransfer(filename, file, p.Remote)

	return nil
}

// resolve file type pack and write the content of pack to file
func (p *Packet) handleFilePacket(s *Service) error {
	realBody := p.Body[protocol.FixedHeaderSize : protocol.FixedHeaderSize+p.proto.PackSize]

	fmt.Printf("[handleFilePacket] receive pack file from %v \n", p.Remote)

	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		return ErrInvalidFilePack
	}

	fmt.Printf("[ORDER] is %d \n", p.proto.PackOrder)

	if p.proto.PackOrder == rem.PackCount {
		fmt.Printf("[Repeat packet] %d \n", p.proto.PackOrder)
		p.Repeat = 1
		return nil
	}

	p.Repeat = 0

	if p.proto.PackOrder-rem.PackCount > 1 {
		return ErrInvalidOrder
	}

	n, err := rem.File.Write(realBody)
	if err != nil {
		return err
	}

	if n < len(realBody) {
		return ErrWrite
	}

	return nil
}

// when file transfer finish, calculate hash of file and compare with hash send by client
func (p *Packet) handleFileFinishPacket(s *Service) error {
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
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
