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
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"os"

	"redalert/protocol"
)

// FileInfo - TCP pack information
type FileInfo struct {
	client     *Client
	replyPack  []byte
	headPack   []byte
	filePack   []byte
	hash       hash.Hash
	file       *os.File
	fileName   os.FileInfo
	fileOffset uint32
}

var (
	errInvalidHeaderSize = errors.New("Header size out of range")
	errWriteIncomplete   = errors.New("Write to tcp not complate")
)

// SendFile send file pack by size
func (fi *FileInfo) SendFile(size int) error {
	n, err := fi.file.Read(fi.filePack[protocol.FixedHeaderSize:])
	if err == io.EOF {
		hashResult := fi.hash.Sum(nil)

		reader := bytes.NewReader(hashResult)
		reader.Read(fi.filePack[protocol.FixedHeaderSize:])
		fi.filePack[0] = protocol.HeaderFileFinishType

		fi.client.conn.Write(fi.headPack)
		fi.client.close <- struct{}{}

		return nil
	}

	if err != nil {
		return err
	}

	fi.hash.Write(fi.filePack[protocol.FixedHeaderSize : protocol.FixedHeaderSize+n])

	binary.LittleEndian.PutUint16(fi.filePack[protocol.PackSizeOffset:], uint16(n))

	_, err = fi.client.conn.Write(fi.filePack)
	if err != nil {
		return err
	}

	return nil
}
