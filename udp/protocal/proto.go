package protocal

const (
	// Header Size
	HeaderTypeSize = 1
	HeaderSize     = 1
	ReplyStatus    = 1
	FileSize       = 8
	PackSize       = 2
	PackCountSize  = 4
	PackOrderSize  = 4
	FileNameSize   = int32(1 << 8)
	RawHeaderSize  = HeaderTypeSize + HeaderSize + FileSize + PackSize + PackCountSize + PackOrderSize + FileNameSize
)
