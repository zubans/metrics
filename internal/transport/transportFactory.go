package transport

import (
	"github.com/zubans/metrics/internal/services"
)

func NewTransport(metricsService *services.MetricsService) TransportInterface {
	if metricsService.Cfg != nil && metricsService.Cfg.UseGRPC {
		return NewGRPCTransport(metricsService)
	}
	return NewHTTPTransport(metricsService)
}
