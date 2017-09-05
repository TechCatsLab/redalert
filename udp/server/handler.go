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
	"fmt"
	"net"
	"redalert/protocol"
	"redalert/udp/remote"
	"time"
)

// Handler represent operations by UDP service
type Handler interface {
	OnError(error, *net.UDPAddr)
	OnPacket(*Packet) error
	OnClose(*Service) error
}

// Provider provide service
type Provider struct{}

var nilPack = make([]byte, 0)

// OnError handle when encounters error
func (sp *Provider) OnError(err error, addr *net.UDPAddr) {
	fmt.Printf("[OnError] crash with error %v \n", err)
	time.Sleep(1 * time.Second)
	remote.Service.Close(addr, err)
}

// OnPacket update client info in the online table according to pack HeaderType
func (sp *Provider) OnPacket(pack *Packet) error {
	fmt.Printf("[OnPacket] pack type is %d, %d, %d \n", pack.proto.HeaderType, len(pack.Body), cap(pack.Body))

	if pack.proto.HeaderType == protocol.HeaderFileFinishType || pack.Repeat == 1 {
		return nil
	}

	if pack.proto.HeaderType == protocol.HeaderRequestType {
		remote.Service.Update(pack.Remote, nilPack)

		return nil
	}

	if pack.proto.HeaderType == protocol.HeaderFileType {
		err := remote.Service.Update(pack.Remote, pack.Body[protocol.FixedHeaderSize:protocol.FixedHeaderSize+pack.proto.PackSize])
		if err != nil {
			return err
		}
	}

	return nil
}

// OnClose close server
func (sp *Provider) OnClose(s *Service) error {
	time.Sleep(1 * time.Second)
	s.Close()

	return nil
}
