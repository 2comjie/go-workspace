package codec

import "errors"

var (
	TypeNotSupportErr error = errors.New("type not support")
	PacketBytesErr    error = errors.New("packet bytes error")
)
