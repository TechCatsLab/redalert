package protocal

const (
	// Header Size
	HeaderTypeSize  = 1
	HeaderSize      = 2
	ReplyStatusSize = 1
	FileSize        = 8
	PackSize        = 2
	PackCountSize   = 4
	PackOrderSize   = 4
	FixedHeaderSize = HeaderTypeSize + HeaderSize + ReplyStatusSize + FileSize + PackSize + PackCountSize + PackOrderSize

	MinRawHeaderSize    = int32(1<<6) - 1
	MiddleRawHeaderSize = int32(1<<7) - 1
	MaxRawHeaderSize    = int32(1<<8) - 1

	MinFileNameSize    = MinRawHeaderSize - FixedHeaderSize
	MiddleFileNameSize = MiddleRawHeaderSize - FixedHeaderSize
	MaxFileNameSize    = MaxRawHeaderSize - FixedHeaderSize

	// Offset
	HeaderTypeOffset  = 0
	HeaderSizeOffset  = HeaderTypeOffset + HeaderTypeSize
	ReplyStatusOffset = HeaderSizeOffset + HeaderSize
	FileSizeOffset    = ReplyStatusOffset + ReplyStatusSize
	PackSizeOffset    = FileSizeOffset + FileSize
	PackCountOffset   = PackSizeOffset + PackSize
	PackOrderOffset   = PackCountOffset + PackCountSize
	FileNameOffset    = PackOrderOffset + PackOrderSize

	// default port
	DefaultUDPPort = "17120"
)

type Proto struct {
	HeaderType  int8  // 包类型 （1.商量协议类型 2.传文件类型）
	HeaderSize  int16 // 包头大小
	ReplyStatus int8  // 返回状态值 （1.yes  2. no）
	FileSize    int64 // 要传输的文件大小
	PackSize    int16 // 包的大小
	PackCount   int32 // 总包量
	PackOrder   int32 // 包序号
}
