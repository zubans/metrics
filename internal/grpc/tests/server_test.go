package grpc_test

import (
	"context"
	"testing"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/grpc"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"github.com/zubans/metrics/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_UpdateMetric(t *testing.T) {
	cfg := &config.Config{
		TrustedSubnet: "",
		GRPCAddr:      "localhost:0",
		EnableGRPC:    true,
	}

	memStorage := storage.NewMemStorage()
	metricService := services.NewMetricService(memStorage)

	server := grpc.NewServer(metricService, cfg)

	req := &proto.UpdateMetricRequest{
		Metric: &proto.Metric{
			Id:   "test_gauge",
			Type: "gauge",
			Value: &proto.Metric_GaugeValue{
				GaugeValue: 123.45,
			},
		},
	}

	resp, err := server.UpdateMetric(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	if resp.Metric == nil {
		t.Fatal("Response metric is nil")
	}

	if resp.Metric.GetId() != "test_gauge" {
		t.Fatalf("Expected metric ID 'test_gauge', got '%s'", resp.Metric.GetId())
	}

	if resp.Metric.GetType() != "gauge" {
		t.Fatalf("Expected metric type 'gauge', got '%s'", resp.Metric.GetType())
	}

	if resp.Metric.GetGaugeValue() != 123.45 {
		t.Fatalf("Expected gauge value 123.45, got %f", resp.Metric.GetGaugeValue())
	}
}

func TestServer_UpdateMetric_InvalidMetric(t *testing.T) {
	cfg := &config.Config{
		TrustedSubnet: "",
		GRPCAddr:      "localhost:0",
		EnableGRPC:    true,
	}

	memStorage := storage.NewMemStorage()
	metricService := services.NewMetricService(memStorage)

	server := grpc.NewServer(metricService, cfg)

	// Test with nil metric
	req := &proto.UpdateMetricRequest{
		Metric: nil,
	}

	_, err := server.UpdateMetric(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for nil metric")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Fatalf("Expected InvalidArgument, got %v", st.Code())
	}
}

func TestServer_UpdateMetrics(t *testing.T) {
	cfg := &config.Config{
		TrustedSubnet: "",
		GRPCAddr:      "localhost:0",
		EnableGRPC:    true,
	}

	memStorage := storage.NewMemStorage()
	metricService := services.NewMetricService(memStorage)

	server := grpc.NewServer(metricService, cfg)

	req := &proto.UpdateMetricsRequest{
		Metrics: []*proto.Metric{
			{
				Id:   "test_gauge_1",
				Type: "gauge",
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 123.45,
				},
			},
			{
				Id:   "test_counter_1",
				Type: "counter",
				Value: &proto.Metric_CounterDelta{
					CounterDelta: 100,
				},
			},
			{
				Id:   "test_gauge_2",
				Type: "gauge",
				Value: &proto.Metric_GaugeValue{
					GaugeValue: 67.89,
				},
			},
		},
	}

	resp, err := server.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateMetrics failed: %v", err)
	}

	if !resp.Success {
		t.Fatal("Expected success response")
	}
}

func TestServer_UpdateMetrics_EmptyMetrics(t *testing.T) {
	cfg := &config.Config{
		TrustedSubnet: "",
		GRPCAddr:      "localhost:0",
		EnableGRPC:    true,
	}

	memStorage := storage.NewMemStorage()
	metricService := services.NewMetricService(memStorage)

	server := grpc.NewServer(metricService, cfg)

	req := &proto.UpdateMetricsRequest{
		Metrics: []*proto.Metric{},
	}

	_, err := server.UpdateMetrics(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for empty metrics")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("Expected gRPC status error")
	}

	if st.Code() != codes.InvalidArgument {
		t.Fatalf("Expected InvalidArgument, got %v", st.Code())
	}
}
