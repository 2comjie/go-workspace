package generic

import "time"

type Number interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint | ~float64 | ~float32 | ~uintptr
}

type Primitive interface {
	Number | ~bool | ~string | ~rune | time.Time
}
