package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

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
	service    handler.ServerMetricService
	cfg        *config.Config
	grpcServer *grpc.Server
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

	s.grpcServer = grpc.NewServer(opts...)
	proto.RegisterMetricsServiceServer(s.grpcServer, s)

	logger.Log.Info("Starting gRPC server", zap.String("address", s.cfg.GRPCAddr))
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() error {
	if s.grpcServer == nil {
		return nil
	}

	logger.Log.Info("Stopping gRPC server...")

	// Graceful shutdown with timeout
	timeout := 10 * time.Second // default timeout
	if s.cfg != nil {
		timeout = s.cfg.ShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		s.grpcServer.GracefulStop()
		done <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Log.Warn("gRPC server graceful shutdown timeout, forcing stop")
		s.grpcServer.Stop()
		return ctx.Err()
	case err := <-done:
		logger.Log.Info("gRPC server stopped gracefully")
		return err
	}
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

	// Convert proto metric to DTO
	dto, err := s.convertProtoToMetricDTO(metric)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convert DTO to MetricData for service layer
	mData, err := s.convertDTOToMetricData(dto)
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

	// Convert proto metrics to DTOs for batch update
	dtoMetrics, err := s.convertProtoMetricsToDTOs(metrics)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Use batch update method from service layer
	_, customErr, err := s.service.UpdateMetrics(ctx, dtoMetrics)
	if err != nil {
		logger.Log.Error("failed to update metrics in batch", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}
	if customErr != nil {
		logger.Log.Error("custom error during batch update", zap.Error(customErr))
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &proto.UpdateMetricsResponse{Success: true}, nil
}

// convertProtoMetricsToDTOs converts slice of proto metrics to DTOs for batch update
func (s *Server) convertProtoMetricsToDTOs(metrics []*proto.Metric) ([]models.MetricsDTO, error) {
	var dtoMetrics []models.MetricsDTO

	for _, metric := range metrics {
		dto, err := s.convertProtoToMetricDTO(metric)
		if err != nil {
			return nil, err
		}
		dtoMetrics = append(dtoMetrics, *dto)
	}

	return dtoMetrics, nil
}

// convertProtoToMetricDTO converts single proto metric to DTO
func (s *Server) convertProtoToMetricDTO(metric *proto.Metric) (*models.MetricsDTO, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	dto := &models.MetricsDTO{
		ID:    metric.GetId(),
		MType: metric.GetType(),
	}

	switch metric.GetType() {
	case "gauge":
		if metric.GetGaugeValue() != 0 {
			val := metric.GetGaugeValue()
			dto.Value = &val
		}
	case "counter":
		if metric.GetCounterDelta() != 0 {
			delta := metric.GetCounterDelta()
			dto.Delta = &delta
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", metric.GetType())
	}

	return dto, nil
}

// convertDTOToMetricData converts DTO to MetricData for service layer
func (s *Server) convertDTOToMetricData(dto *models.MetricsDTO) (*services.MetricData, error) {
	if dto == nil {
		return nil, fmt.Errorf("dto is nil")
	}

	mData := &services.MetricData{
		Type: dto.MType,
		Name: dto.ID,
	}

	switch dto.MType {
	case "gauge":
		if dto.Value != nil {
			val := strconv.FormatFloat(*dto.Value, 'f', -1, 64)
			mData.Value = &val
		}
	case "counter":
		if dto.Delta != nil {
			delta := strconv.FormatInt(*dto.Delta, 10)
			mData.Value = &delta
		}
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", dto.MType)
	}

	return mData, nil
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
		if fval, err := strconv.ParseFloat(value, 64); err == nil {
			metric.Value = &proto.Metric_GaugeValue{GaugeValue: fval}
		} else {
			logger.Log.Warn("failed to parse gauge value", zap.String("value", value), zap.Error(err))
		}
	} else if req.GetType() == "counter" {
		if ival, err := strconv.ParseInt(value, 10, 64); err == nil {
			metric.Value = &proto.Metric_CounterDelta{CounterDelta: ival}
		} else {
			logger.Log.Warn("failed to parse counter value", zap.String("value", value), zap.Error(err))
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
		val := strconv.FormatFloat(metric.GetGaugeValue(), 'f', -1, 64)
		mData.Value = &val
	case "counter":
		val := strconv.FormatInt(metric.GetCounterDelta(), 10)
		mData.Value = &val
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
