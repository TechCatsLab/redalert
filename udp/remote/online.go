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
	"crypto/md5"
	"fmt"
	"hash"
	"io"
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
	PackCount uint32
	Timer     *time.Timer
	Hash      hash.Hash
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
func (r *remoteAddrTable) OnStartTransfor(filename string, file *os.File, remote *net.UDPAddr) bool {
	_, ok := r.remote[remote]
	if !ok {
		rem := Remote{
			FileName: filename,
			File:     file,
			Timer: time.AfterFunc(3*time.Minute, func() {
				r.Close(remote)
			}),
			Hash: md5.New(),
		}

		r.remote[remote] = &rem

		fmt.Printf("[OnStartTransfor] create a table %v \n", *remote)
		return true
	}

	return false
}

// GetRemote return *Remote and true if exists
func (r *remoteAddrTable) GetRemote(rmt *net.UDPAddr) (*Remote, bool) {
	fmt.Printf("[GetRemote] quering %v \n", rmt)
	rem, ok := r.remote[rmt]
	fmt.Printf("get %v \n", rem)
	for k, v := range r.remote {
		if rmt == k {
			fmt.Printf("true \n")
		} else {

			fmt.Printf("ip address is %v and contant is %v \n", k, v)
		}
	}
	return rem, ok
}

// Update update timer and count when receive success
func (r *remoteAddrTable) Update(remote *net.UDPAddr, pack *[]byte) {
	fmt.Printf("[Update] map \n")
	rem, _ := r.remote[remote]

	fmt.Printf("rem is %v \n", rem)

	rem.PackCount++
	rem.Timer.Reset(3 * time.Minute)
	io.WriteString(rem.Hash, string(*pack))

	rr, _ := r.remote[remote]
	fmt.Printf("now count is %d \n", rr.PackCount)

	// r.remote[remote] = rem
}

// Close file and delete map
func (r *remoteAddrTable) Close(remote *net.UDPAddr) {
	fmt.Printf("[Close] remote %v \n", remote)
	rem, ok := r.remote[remote]
	if !ok {
		return
	}

	rem.File.Close()
	os.Remove(rem.FileName)
	rem.Timer.Stop()
	delete(r.remote, remote)
}

func (r *remoteAddrTable) ResetTimer(addr *net.UDPAddr) {
	fmt.Printf("[ResetTimer] Executing reset timer func \n")
	rem, _ := r.remote[addr]

	rem.Timer.Reset(3 * time.Minute)
}
