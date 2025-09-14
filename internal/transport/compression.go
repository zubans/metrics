package transport

import (
	"bytes"
	"compress/gzip"
)

type CompressionService struct{}

func NewCompressionService() *CompressionService {
	return &CompressionService{}
}

func (cs *CompressionService) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (cs *CompressionService) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Encoding": "gzip",
	}
}

func (cs *CompressionService) IsCompressionEnabled() bool {
	return true
}
