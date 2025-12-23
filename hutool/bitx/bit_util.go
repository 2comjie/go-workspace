package bitx

// go:force inline
func SetBit(b *byte, pos uint, val bool) {
	if b == nil || pos > 7 { // 增加指针空判断
		return
	}
	if val {
		*b |= 1 << pos
	} else {
		*b &^= 1 << pos
	}
}

// go:force inline
func ClearBit(b *byte, pos uint) {
	if b == nil || pos > 7 {
		return
	}
	*b &^= 1 << pos
}

// go:force inline
func IsBitSet(b byte, pos uint) bool {
	if pos > 7 {
		return false
	}
	return (b & (1 << pos)) != 0
}
