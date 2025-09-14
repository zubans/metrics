package transport

import (
	"log"

	"github.com/zubans/metrics/internal/services"
)

type DataProcessor struct {
	compressionService *CompressionService
	encryptionService  *EncryptionService
}

func NewDataProcessor(metricsService *services.MetricsService) *DataProcessor {
	return &DataProcessor{
		compressionService: NewCompressionService(),
		encryptionService:  NewEncryptionService(metricsService),
	}
}

func (dp *DataProcessor) ProcessData(data []byte) (interface{}, map[string]string, error) {
	compressedData, err := dp.compressionService.Compress(data)
	if err != nil {
		log.Printf("Error compressing data: %v", err)
		return nil, nil, err
	}

	encryptedData, extraHeaders, err := dp.encryptionService.EncryptData(compressedData)
	if err != nil {
		log.Printf("Error encrypting data: %v", err)
		return nil, nil, err
	}

	headers := dp.compressionService.GetHeaders()
	for k, v := range extraHeaders {
		headers[k] = v
	}

	return encryptedData, headers, nil
}

func (dp *DataProcessor) GetHeaders() map[string]string {
	headers := dp.compressionService.GetHeaders()

	if dp.encryptionService.IsEncryptionEnabled() {
		delete(headers, "Content-Encoding")
		headers["X-Encrypted"] = "1"
	}

	return headers
}
