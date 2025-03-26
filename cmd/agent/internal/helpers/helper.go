package helpers

import (
	"fmt"
	"github.com/zubans/metrics/cmd/agent/internal/models"
)

func ToURL(m models.Metric) string {
	return fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", m.Type, m.Name, m.Value)
}
