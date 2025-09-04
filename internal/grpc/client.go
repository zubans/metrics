package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client proto.MetricsServiceClient
	cfg    *config.AgentConfig
}

func NewClient(cfg *config.AgentConfig) (*Client, error) {
	conn, err := grpc.Dial(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SendMetrics(metrics *models.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var protoMetrics []*proto.Metric
	for _, metric := range metrics.MetricList {
		protoMetric := c.convertToProtoMetric(metric)
		protoMetrics = append(protoMetrics, protoMetric)
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	}

	resp, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("server returned failure")
	}

	log.Printf("Successfully sent %d metrics via gRPC", len(protoMetrics))
	return nil
}

func (c *Client) SendSingleMetric(metric models.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	protoMetric := c.convertToProtoMetric(metric)

	req := &proto.UpdateMetricRequest{
		Metric: protoMetric,
	}

	_, err := c.client.UpdateMetric(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metric %s: %w", metric.Name, err)
	}

	log.Printf("Successfully sent metric %s via gRPC", metric.Name)
	return nil
}

func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &proto.PingRequest{}
	resp, err := c.client.Ping(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("server ping failed")
	}

	return nil
}

func (c *Client) convertToProtoMetric(metric models.Metric) *proto.Metric {
	protoMetric := &proto.Metric{
		Id:   metric.Name,
		Type: string(metric.Type),
	}

	switch metric.Type {
	case models.Gauge:
		protoMetric.Value = &proto.Metric_GaugeValue{
			GaugeValue: float64(metric.Value),
		}
	case models.Counter:
		protoMetric.Value = &proto.Metric_CounterDelta{
			CounterDelta: int64(metric.Value),
		}
	}

	return protoMetric
}
