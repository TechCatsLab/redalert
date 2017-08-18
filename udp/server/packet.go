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
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
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

	return err
}

func (p *Packet) Read(s *Service) error {
	size, remote, err := s.conn.ReadFromUDP(p.Body)

	if err != nil {
		return err
	}

	p.Size = size
	p.Remote = remote

	pack := Packet{
		Remote: p.Remote,
		Size:   1,
	}

	headerType := uint8(p.Body[0])
	switch {
	case headerType == protocal.HeaderRequestType:
		err := p.handleRequest(s)
		if err != nil {
			pack.Body[0] = protocal.ReplyNo
			err := pack.WriteToUDP(s.conn)
			if err != nil {
				return err
			}
		} else {
			pack.Body[0] = protocal.ReplyOk
			err := pack.WriteToUDP(s.conn)
			if err != nil {
				return err
			}
		}
	case headerType == protocal.HeaderFileType:
		err := p.handleFilePacket(s)
		if err != nil {
			pack.Body[0] = protocal.ReplyNo
			err := pack.WriteToUDP(s.conn)
			if err != nil {
				return err
			}
		} else {
			pack.Body[0] = protocal.ReplyOk
			err := pack.WriteToUDP(s.conn)
			if err != nil {
				return err
			}
		}
	case headerType == protocal.HeaderFileFinishType:
		err := p.handleFileFinishPacket(s, "hash from client")
		if err != nil {
			return err
		}
	}

	return nil
}

// Reset the buffer
func (p *Packet) Reset() {
	p.Size = 0
	p.Remote = nil
}

func (p *Packet) handleRequest(s *Service) (err error) {
	binary.LittleEndian.PutUint16(p.Body[protocal.HeaderSizeOffset:protocal.FileSizeOffset], p.proto.HeaderSize)
	filename := string(p.Body[protocal.FileNameOffset : p.proto.HeaderSize-protocal.FixedHeaderSize])

	if rem, ok := remote.Service.GetRemote(p.Remote); ok {
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

	if ok := remote.Service.OnStartTransfor(filename, file, p.Remote); !ok {
		return ErrDiffrentFile
	}

	return nil
}

func (p *Packet) handleFilePacket(s *Service) error {
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		return ErrInvalidFilePack
	}

	binary.LittleEndian.PutUint32(p.Body[protocal.PackOrderOffset:protocal.FixedHeaderSize], p.proto.PackOrder)
	if p.proto.PackOrder-rem.PackCount > 1 {
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

func (p *Packet) handleFileFinishPacket(s *Service, h string) error {
	rem, ok := remote.Service.GetRemote(p.Remote)
	if !ok {
		return ErrNotExists
	}

	hash, err := checkHash(rem.File)
	if err != nil {
		return err
	}

	if hash != h {
		return ErrHashNotMatch
	}

	return nil
}

// todo: 考虑大文件哈希
func checkHash(file *os.File) (string, error) {
	md5h := md5.New()
	_, err := io.Copy(md5h, file)
	if err != nil {
		return "", err
	}

	return string(md5h.Sum(nil)), nil
}
