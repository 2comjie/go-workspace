package zip

import "hutool/zip"

type GZIP struct {
}

func (G GZIP) Zip(data []byte) ([]byte, error) {
	return zip.GzipCompress(data)
}

func (G GZIP) Unzip(data []byte) ([]byte, error) {
	return zip.GzipDecompress(data)
}
