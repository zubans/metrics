package storage

import (
	"context"
	"encoding/json"
	"github.com/zubans/metrics/internal/config"
	"log"
	"os"
)

type PersistentStorage interface {
	SaveMetricToFile() error
	LoadMetricsFromFile() error
}

type Dump struct {
	storage *MemStorage
	cfg     *config.Config
}

type MetricsDump struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func NewDump(storage *MemStorage, cfg config.Config) *Dump {
	return &Dump{storage: storage, cfg: &cfg}
}

func (f *Dump) SaveMetricToFile(ctx context.Context) error {
	dump := MetricsDump{
		Gauges:   f.storage.GetGauges(ctx),
		Counters: f.storage.GetCounters(ctx),
	}

	data, err := json.Marshal(dump)
	if err != nil {
		return err
	}

	return os.WriteFile(f.cfg.FileStoragePath, data, 0644)
}

func (f *Dump) LoadMetricsFromFile() error {
	res, err := os.ReadFile(f.cfg.FileStoragePath)
	if err != nil {
		log.Println("error open file")
	}

	err = json.Unmarshal(res, f.storage)
	if err != nil {
		return err
	}

	return nil
}
