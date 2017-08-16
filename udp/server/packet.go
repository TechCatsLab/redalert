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
	"net"
	"os"

	"redalert/udp/protocal"
)

// Packet represent a UDP packet
type Packet struct {
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

	return err
}

func (p *Packet) Read(conn *net.UDPConn) error {
	size, remote, err := conn.ReadFromUDP(p.Body)

	if err != nil {
		return err
	}

	p.Size = size
	p.Remote = remote

	return nil
}

// Reset the buffer
func (p *Packet) Reset() {
	p.Size = 0
	p.Remote = nil
}

func (p *Packet) handleRequest(s *Service) (err error) {
	headerSize := protocal.Int16(p.Body[protocal.HeaderSizeOffset:protocal.FileSizeOffset])
	filename := string(p.Body[protocal.FileNameOffset : headerSize-protocal.FixedHeaderSize])

	if rem, ok := s.remote[p.Remote]; ok {
		if filename != rem.FileName {
			err = ErrDiffrentFile
			return
		}

		err = ErrDuplicated
		return
	}

	file, err := os.OpenFile(protocal.DefaultDir+filename, 0, 0666)
	if err != nil {
		return
	}

	s.remote[p.Remote] = &Remote{
		FileName: filename,
		File:     file,
	}

	return nil
}

func (p *Packet) handleFilePacket(s *Service) error {
	rem, ok := checkpack(p.Remote, s)
	if !ok {
		return ErrInvalidFilePack
	}

	packOrder := protocal.Int32(p.Body[protocal.PackOrderOffset:protocal.FixedHeaderSize])
	if packOrder-rem.PackCount > 1 {
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

func checkpack(addr *net.UDPAddr, s *Service) (*Remote, bool) {
	remote, ok := s.remote[addr]

	return remote, ok
}
