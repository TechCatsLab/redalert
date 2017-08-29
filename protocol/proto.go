/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co.,Ltd..
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
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
 * Revision History:
 *     Initial: 2017/08/10        Yusan Kurban
 *     Initial: 2017/08/10        Liu Jiachang
 */

package protocol

const (
	// HeaderTypeSize - Header Size
	HeaderTypeSize  = 1
	HeaderSize      = 2
	FileSize        = 8
	PackSize        = 2
	PackCountSize   = 4
	PackOrderSize   = 4
	FixedHeaderSize = HeaderTypeSize + HeaderSize + FileSize + PackSize + PackCountSize + PackOrderSize

	ReplySize = 4

	RawHeaderSize    = int32(1<<6) - 1
	ReqRawHeaderSize = int32(1<<8) - 1

	ReqFileNameSize = ReqRawHeaderSize - FixedHeaderSize

	// HeaderTypeOffset - Offset
	HeaderTypeOffset = 0
	HeaderSizeOffset = HeaderTypeOffset + HeaderTypeSize
	FileSizeOffset   = HeaderSizeOffset + HeaderSize
	PackSizeOffset   = FileSizeOffset + FileSize
	PackCountOffset  = PackSizeOffset + PackSize
	PackOrderOffset  = PackCountOffset + PackCountSize
	FileNameOffset   = PackOrderOffset + PackOrderSize

	// DefaultUDPPort - default port
	DefaultUDPPort = "17120"

	// HeaderRequestType - define HeaderType
	HeaderRequestType    = 0x10
	HeaderFileType       = 0x20
	HeaderFileFinishType = 0x30

	// ReplyFinish - define Reply finish
	ReplyFinish = 1<<32 - 1
	ReplyError  = 1<<32 - 2

	RepeatHandle = 1<<32 - 3

	// DefaultDir is default dir for save file
	DefaultDir = "./"

	FirstPacketSize = 1 << 9
	MaxPacketSize   = 1 << 13
)

// Proto - 传输协议结构
type Proto struct {
	HeaderType uint8  // 包类型 （1.商量协议类型 2.传文件类型）
	HeaderSize uint16 // 第一包整个包大小，以后包头大小
	FileSize   uint64 // 要传输的文件大小
	PackSize   uint16 // 第一包传以后每个包的大小
	PackCount  uint32 // 总包量
	PackOrder  uint32 // 包序号
}
