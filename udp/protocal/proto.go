package protocal

const (
	// Header Size
	HeaderTypeSize  = 1 // 包类型
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
