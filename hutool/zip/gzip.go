package zip

import (
	"bytes"
	"compress/gzip"
	"io"
)

func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = gzReader.Close()
	}()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gzReader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
