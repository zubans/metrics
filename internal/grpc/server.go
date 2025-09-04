package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/models"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedMetricsServiceServer
	service handler.ServerMetricService
	cfg     *config.Config
}

func NewServer(service handler.ServerMetricService, cfg *config.Config) *Server {
	return &Server{
		service: service,
		cfg:     cfg,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.cfg.GRPCAddr, err)
	}

	var opts []grpc.ServerOption
	if s.cfg.TrustedSubnet != "" {
		opts = append(opts, grpc.UnaryInterceptor(s.trustedSubnetInterceptor))
	}

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterMetricsServiceServer(grpcServer, s)

	logger.Log.Info("Starting gRPC server", zap.String("address", s.cfg.GRPCAddr))
	return grpcServer.Serve(lis)
}

func (s *Server) trustedSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get peer info")
	}

	if !s.isIPInTrustedSubnet(p.Addr.String()) {
		logger.Log.Warn("gRPC request from untrusted IP", zap.String("peer", p.Addr.String()))
		return nil, status.Error(codes.PermissionDenied, "request from untrusted subnet")
	}

	return handler(ctx, req)
}

func (s *Server) isIPInTrustedSubnet(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	_, ipnet, err := net.ParseCIDR(s.cfg.TrustedSubnet)
	if err != nil {
		return false
	}

	return ipnet.Contains(ip)
}

func (s *Server) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	metric := req.GetMetric()
	if metric == nil {
		return nil, status.Error(codes.InvalidArgument, "metric is required")
	}

	mData, err := s.convertProtoToMetricData(metric)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	updatedMetric, _, err := s.service.UpdateMetric(ctx, mData)
	if err != nil {
		logger.Log.Error("failed to update metric", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update metric")
	}

	responseMetric := s.convertMetricDataToProto(updatedMetric)
	return &proto.UpdateMetricResponse{Metric: responseMetric}, nil
}

func (s *Server) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	metrics := req.GetMetrics()
	if len(metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one metric is required")
	}

	var dtoMetrics []interface{}
	for _, metric := range metrics {
		mData, err := s.convertProtoToMetricData(metric)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		dtoMetrics = append(dtoMetrics, mData)
	}

	// TODO: Implement batch update in service layer
	for _, dtoMetric := range dtoMetrics {
		_, _, err := s.service.UpdateMetric(ctx, dtoMetric.(*services.MetricData))
		if err != nil {
			logger.Log.Error("failed to update metric in batch", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to update metrics")
		}
	}

	return &proto.UpdateMetricsResponse{Success: true}, nil
}

func (s *Server) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	if req.GetId() == "" || req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, "id and type are required")
	}

	mData := &services.MetricData{
		Type: req.GetType(),
		Name: req.GetId(),
	}

	value, details := s.service.GetMetric(ctx, mData)
	if details != nil {
		return nil, status.Error(codes.NotFound, "metric not found")
	}

	metric := &proto.Metric{
		Id:   req.GetId(),
		Type: req.GetType(),
	}

	if req.GetType() == "gauge" {
		if val, err := fmt.Sscanf(value, "%f", new(float64)); err == nil && val == 1 {
			var fval float64
			fmt.Sscanf(value, "%f", &fval)
			metric.Value = &proto.Metric_GaugeValue{GaugeValue: fval}
		}
	} else if req.GetType() == "counter" {
		if val, err := fmt.Sscanf(value, "%d", new(int64)); err == nil && val == 1 {
			var ival int64
			fmt.Sscanf(value, "%d", &ival)
			metric.Value = &proto.Metric_CounterDelta{CounterDelta: ival}
		}
	}

	return &proto.GetMetricResponse{Metric: metric}, nil
}

func (s *Server) ListMetrics(ctx context.Context, req *proto.ListMetricsRequest) (*proto.ListMetricsResponse, error) {
	_, err := s.service.ShowMetrics(ctx)
	if err != nil {
		logger.Log.Error("failed to get metrics", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get metrics")
	}

	// TODO: Implement proper list method in service layer
	return &proto.ListMetricsResponse{Metrics: []*proto.Metric{}}, nil
}

func (s *Server) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	err := s.service.Ping(ctx)
	if err != nil {
		return &proto.PingResponse{Success: false}, nil
	}
	return &proto.PingResponse{Success: true}, nil
}

func (s *Server) convertProtoToMetricData(metric *proto.Metric) (*services.MetricData, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	mData := &services.MetricData{
		Type: metric.GetType(),
		Name: metric.GetId(),
	}

	switch metric.GetType() {
	case "gauge":
		if metric.GetGaugeValue() != 0 {
			val := fmt.Sprintf("%f", metric.GetGaugeValue())
			mData.Value = &val
		}
	case "counter":
		if metric.GetCounterDelta() != 0 {
			val := fmt.Sprintf("%d", metric.GetCounterDelta())
			mData.Value = &val
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metric.GetType())
	}

	return mData, nil
}

func (s *Server) convertMetricDataToProto(metric *models.MetricsDTO) *proto.Metric {
	protoMetric := &proto.Metric{
		Id:   metric.ID,
		Type: metric.MType,
	}

	switch metric.MType {
	case "gauge":
		if metric.Value != nil {
			protoMetric.Value = &proto.Metric_GaugeValue{GaugeValue: *metric.Value}
		}
	case "counter":
		if metric.Delta != nil {
			protoMetric.Value = &proto.Metric_CounterDelta{CounterDelta: *metric.Delta}
		}
	}

	return protoMetric
}
