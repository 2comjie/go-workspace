package zip

type IZip interface {
	Zip(data []byte) ([]byte, error)
	Unzip(data []byte) ([]byte, error)
}
