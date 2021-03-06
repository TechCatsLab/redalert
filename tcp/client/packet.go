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
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/TechCatsLab/redalert/protocol"
)

// FileInfo - TCP pack information
type FileInfo struct {
	client     *Client
	replyPack  []byte
	headPack   []byte
	filePack   []byte
	hash       hash.Hash
	file       *os.File
	fileInfo   os.FileInfo
	fileOffset uint32
}

var (
	errInvalidHeaderSize = errors.New("Header size out of range")
	errFromServer        = errors.New("Got error from server")
	errPackOrder         = errors.New("Pack order messed")
)

func (fi *FileInfo) initFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(name)
	if err != nil {
		return err
	}

	fmt.Printf("file name is %v size is %v: \n", fileInfo.Name(), fileInfo.Size())
	fi.file = file
	fi.fileInfo = fileInfo
	fi.hash = md5.New()

	return nil
}

// first pack which for consult
func (fi *FileInfo) consult() error {
	fi.client.proto.PackSize = uint16(fi.client.conf.PackSize)
	fi.client.proto.HeaderSize = uint16(len(fi.fileInfo.Name()) + protocol.FixedHeaderSize)
	fi.headPack = make([]byte, protocol.FirstPacketSize)

	err := fi.packHead(fi.headPack)
	if err != nil {
		return err
	}

	fi.headPack[0] = byte(protocol.HeaderRequestType)
	nameReader := strings.NewReader(fi.fileInfo.Name())
	nameReader.Read(fi.headPack[protocol.FixedHeaderSize:])

	n, err := fi.client.conn.Write(fi.headPack)
	if err != nil {
		fi.client.handle.OnError(err)
	}

	fmt.Printf("[WRITE] write %d byte \n", n)

	return nil
}

func (fi *FileInfo) packHead(b []byte) error {
	if len(b) < protocol.FixedHeaderSize {
		return errInvalidHeaderSize
	}

	firstPack := protocol.Encode{
		Body: b,
	}
	firstPack.Buffer = bytes.NewBuffer(firstPack.Body)

	// marshal
	firstPack.Marshal(fi.client.proto)

	return nil
}

// SendFile send file pack by size
func (fi *FileInfo) SendFile(size int) error {
	n, err := fi.file.Read(fi.filePack[protocol.FixedHeaderSize:])
	if err != nil {
		if err == io.EOF {
			hashResult := fi.hash.Sum(nil)

			fmt.Printf("[DEBUG]:Send file finish.hash %x\n", hashResult)

			reader := bytes.NewReader(hashResult)
			reader.Read(fi.filePack[protocol.FixedHeaderSize:])
			fi.filePack[0] = protocol.HeaderFileFinishType

			_, err = fi.client.conn.Write(fi.filePack[:protocol.FixedHeaderSize+16])
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	_, err = fi.hash.Write(fi.filePack[protocol.FixedHeaderSize:protocol.FixedHeaderSize+n])
	if err != nil {
		return err
	}

	fi.client.proto.PackOrder++

	filePacket := protocol.Encode{
		Body: fi.filePack,
	}
	filePacket.Buffer = bytes.NewBuffer(filePacket.Body)

	filePacket.Marshal(fi.client.proto)
	fi.filePack[0] = protocol.HeaderFileType

	fmt.Printf("[SendFile] send pack order is %v \n", fi.client.proto.PackOrder)
	_, err = fi.client.conn.Write(fi.filePack[:protocol.FixedHeaderSize+n])
	if err != nil {
		return err
	}

	return nil
}
