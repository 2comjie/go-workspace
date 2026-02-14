package zip

type None struct {
}

func (n None) Zip(data []byte) ([]byte, error) {
	return data, nil
}

func (n None) Unzip(data []byte) ([]byte, error) {
	return data, nil
}
