package helpers

import (
	"fmt"
	"github.com/zubans/metrics/cmd/agent/internal/config"
	"github.com/zubans/metrics/cmd/agent/internal/models"
)

func ToURL(m models.Metric, cfg config.Config) string {
	return fmt.Sprintf("http://%s/update/%s/%s/%d", cfg.AddressServer, m.Type, m.Name, m.Value)
}
