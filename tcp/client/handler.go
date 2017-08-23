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
*     Initial: 2017/08/23          Yusan Kurban
 */

package client

import (
	"crypto/md5"
	"fmt"
	"hash"
	"os"
)

type handler interface {
	OnError(error)
	OnClose() error
	OnReceive(int) error
	OnSend(int) error
}

// DefaultHandler - TCP pack handler
type DefaultHandler struct {
	replyPack []byte
	sendPack  []byte
	hash      hash.Hash
	file      *os.File
	FileName  os.FileInfo
}

func (h *DefaultHandler) request(name string) error {
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("[ERROR] Open file crash with error: %v \n", err)
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("[ERROR] Get file info crash with error: %v \n", err)
		return err
	}

	h.file = file
	h.FileName = fileInfo
	h.hash = md5.New()

	return nil
}
