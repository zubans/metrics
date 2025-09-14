package transport

import (
	"crypto/rsa"
	"log"

	"github.com/zubans/metrics/internal/cryptoutil"
	"github.com/zubans/metrics/internal/services"
)

type EncryptionService struct {
	publicKey *rsa.PublicKey
}

func NewEncryptionService(metricsService *services.MetricsService) *EncryptionService {
	es := &EncryptionService{}

	if metricsService.Cfg != nil && metricsService.Cfg.CryptoKey != "" {
		if pub, err := cryptoutil.LoadPublicKey(metricsService.Cfg.CryptoKey); err == nil {
			es.publicKey = pub
		} else {
			log.Printf("failed to load public key: %v", err)
		}
	}

	return es
}

func (es *EncryptionService) EncryptData(data []byte) (interface{}, map[string]string, error) {
	extraHeaders := map[string]string{}

	if es.publicKey != nil {
		env, encErr := cryptoutil.EncryptHybrid(es.publicKey, data)
		if encErr != nil {
			return nil, nil, encErr
		}
		extraHeaders["X-Encrypted"] = "1"
		extraHeaders["Content-Encoding"] = "" // Убираем gzip заголовок при шифровании
		return env, extraHeaders, nil
	}

	return data, extraHeaders, nil
}

func (es *EncryptionService) IsEncryptionEnabled() bool {
	return es.publicKey != nil
}
