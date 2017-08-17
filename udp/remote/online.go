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
*     Initial: 2017/08/17          Yusan Kurban
 */

package remote

import (
	"net"
	"os"
	"time"
)

// Service expose interface of RemoteAddrTable
var Service *remoteAddrTable

// Remote storage remote client info
type Remote struct {
	FileName  string
	File      *os.File
	PackCount int32
	Timer     *time.Timer
}

// RemoteAddrTable manege remote client address and it's transformation info
type remoteAddrTable struct {
	remote map[*net.UDPAddr]*Remote
}

func init() {
	Service = &remoteAddrTable{
		remote: make(map[*net.UDPAddr]*Remote),
	}
}

// OnStartTransfor storage Remote for new client
func (r *remoteAddrTable) OnStartTransfor(filename string, file *os.File, count int32, remote *net.UDPAddr) bool {
	_, ok := r.remote[remote]
	if !ok {
		rem := Remote{
			FileName:  filename,
			File:      file,
			PackCount: count,
			Timer: time.AfterFunc(3*time.Minute, func() {
				r.Remove(remote)
			}),
		}

		r.remote[remote] = &rem

		return true
	}

	return false
}

// GetRemote return *Remote and true if exists
func (r *remoteAddrTable) GetRemote(remote *net.UDPAddr) (*Remote, bool) {
	rem, ok := r.remote[remote]

	return rem, ok
}

// Remove remove the disconnect client
func (r *remoteAddrTable) Remove(remote *net.UDPAddr) {
	delete(r.remote, remote)
}
