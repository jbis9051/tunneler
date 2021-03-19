package internal

const TunnelerVersion = 1

const (
	SuccessResponse              = uint8(0)
	AddrTypeNotSupportedResponse = uint8(1)
	RuleFailureResponse          = uint8(2)
	ConnectionErrorResponse      = uint8(3)
	UnknownErrorResponse         = uint8(4)
)
