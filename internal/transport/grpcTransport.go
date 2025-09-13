package transport

import (
	"log"

	"github.com/zubans/metrics/internal/grpc"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
)

type GRPCTransport struct {
	grpcClient *grpc.Client
}

func NewGRPCTransport(metricsService *services.MetricsService) *GRPCTransport {
	var grpcClient *grpc.Client
	if metricsService.Cfg != nil && metricsService.Cfg.UseGRPC {
		if client, err := grpc.NewClient(metricsService.Cfg); err == nil {
			grpcClient = client
		} else {
			log.Printf("failed to create gRPC client: %v", err)
		}
	}

	return &GRPCTransport{
		grpcClient: grpcClient,
	}
}

func (gt *GRPCTransport) SendMetrics(metrics []models.Metric) error {
	if gt.grpcClient == nil {
		return nil
	}

	metricsList := &models.Metrics{
		MetricList: metrics,
	}

	if err := gt.grpcClient.SendMetrics(metricsList); err != nil {
		log.Printf("Error sending metrics via gRPC: %v", err)
		return err
	}
	return nil
}

func (gt *GRPCTransport) SendSingleMetric(metric models.Metric) error {
	if gt.grpcClient == nil {
		return nil
	}

	if err := gt.grpcClient.SendSingleMetric(metric); err != nil {
		log.Printf("Error sending metric %s via gRPC: %v", metric.Name, err)
		return err
	}
	return nil
}

func (gt *GRPCTransport) Close() error {
	if gt.grpcClient != nil {
		if err := gt.grpcClient.Close(); err != nil {
			log.Printf("Error closing gRPC client: %v", err)
			return err
		}
	}
	return nil
}
