package internal

const TunnelerVersion = 1

const (
	SuccessResponse              = uint8(0)
	AddrTypeNotSupportedResponse = uint8(1)
	ConnectionErrorResponse      = uint8(2)
	UnknownErrorResponse         = uint8(3)
)
