package iox

import (
	"io"
	"math"
)

func ReadFixBytes(r io.Reader, b []byte) error {
	// 读取固定长度
	_, err := io.ReadFull(r, b)
	return err
}

func WriteLimit(writer io.Writer, bytes []byte, limit int32) error {
	start := 0
	maxLen := len(bytes)
	for {
		currentWriteLen := int(math.Min(float64(limit), float64(maxLen-start)))
		writeLen, writeErr := writer.Write(bytes[start : start+currentWriteLen])
		if writeErr != nil {
			return writeErr
		}
		start += writeLen
		if start >= maxLen {
			break
		}
	}
	return nil
}
