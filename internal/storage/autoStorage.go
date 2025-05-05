package storage

import (
	"context"
	"fmt"
	"github.com/zubans/metrics/internal/models"
	"log"
)

type AutoStorage struct {
	storage *MemStorage
	dump    *Dump
}

func NewAutoDump(storage *MemStorage, dump *Dump) *AutoStorage {
	return &AutoStorage{storage: storage, dump: dump}
}

func (s *AutoStorage) UpdateGauge(ctx context.Context, name string, value float64) float64 {
	res := s.dump.storage.UpdateGauge(ctx, name, value)
	err := s.dump.SaveMetricToFile(ctx)
	if err != nil {
		log.Println("error save gauge to file")
	}

	return res
}

func (s *AutoStorage) UpdateCounter(ctx context.Context, name string, value int64) int64 {
	res := s.dump.storage.UpdateCounter(ctx, name, value)
	err := s.dump.SaveMetricToFile(ctx)
	if err != nil {
		log.Println("error save counter to file")
	}

	return res
}

func (s *AutoStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	return s.dump.storage.GetGauge(ctx, name)
}
func (s *AutoStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	return s.dump.storage.GetCounter(ctx, name)
}
func (s *AutoStorage) GetGauges(ctx context.Context) map[string]float64 {
	return s.dump.storage.GetGauges(ctx)
}
func (s *AutoStorage) GetCounters(ctx context.Context) map[string]int64 {
	return s.dump.storage.GetCounters(ctx)
}
func (s *AutoStorage) ShowMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	return s.dump.storage.ShowMetrics(ctx)
}

func (s *AutoStorage) UpdateMetrics(ctx context.Context, m []models.MetricsDTO) error {
	return fmt.Errorf("forbidden")
}
