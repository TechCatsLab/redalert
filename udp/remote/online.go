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
	"errors"
	"fmt"
	"hash"
	"net"
	"os"
	"time"
)

// Service expose interface of RemoteAddrTable
var (
	Service    *remoteAddrTable
	errTimeOut = errors.New("receive time out")
)

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
	remote map[string]*Remote
}

func init() {
	Service = &remoteAddrTable{
		remote: make(map[string]*Remote),
	}
}

// OnStartTransfor storage Remote for new client
func (r *remoteAddrTable) OnStartTransfor(filename string, file *os.File, remote *net.UDPAddr) bool {
	key := remote.IP.String() + ":" + string(remote.Port)
	_, ok := r.remote[key]
	if !ok {
		rem := Remote{
			FileName: filename,
			File:     file,
			Timer: time.AfterFunc(3*time.Minute, func() {
				r.Close(remote, errTimeOut)
			}),
			Hash: md5.New(),
		}

		r.remote[key] = &rem

		fmt.Printf("[OnStartTransfor] create a table %v \n", remote)
		return true
	}

	return false
}

// GetRemote return *Remote and true if exists
func (r *remoteAddrTable) GetRemote(rmt *net.UDPAddr) (*Remote, bool) {
	key := rmt.IP.String() + ":" + string(rmt.Port)
	fmt.Printf("[GetRemote] quering %v \n", rmt)
	rem, ok := r.remote[key]

	return rem, ok
}

// Update update timer and count when receive success
func (r *remoteAddrTable) Update(remote *net.UDPAddr, pack []byte) error {
	key := remote.IP.String() + ":" + string(remote.Port)
	fmt.Printf("[Update] map \n")
	rem, _ := r.remote[key]

	rem.PackCount++
	rem.Timer.Reset(3 * time.Minute)
	_, err := rem.Hash.Write(pack)
	if err != nil {
		return err
	}

	rr, _ := r.remote[key]
	fmt.Printf("[Update] now count is %d \n", rr.PackCount)

	return nil
	// r.remote[remote] = rem
}

// Close file and delete map
func (r *remoteAddrTable) Close(remote *net.UDPAddr, err error) {
	key := remote.IP.String() + ":" + string(remote.Port)
	fmt.Printf("[Close] remote %v \n", remote)
	rem, ok := r.remote[key]
	if !ok {
		return
	}

	rem.File.Close()
	os.Remove(rem.FileName)
	rem.Timer.Stop()
	if err != nil {
		delete(r.remote, key)
	}
}

func (r *remoteAddrTable) ResetTimer(addr *net.UDPAddr) {
	key := addr.IP.String() + ":" + string(addr.Port)
	fmt.Printf("[ResetTimer] Executing reset timer func \n")
	rem, _ := r.remote[key]

	rem.Timer.Reset(3 * time.Minute)
}
