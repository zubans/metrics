package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zubans/metrics/internal/config"
)

const (
	maxRetries = 3
)

var retryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// PersistentStorage определяет интерфейс для сохранения и загрузки метрик из файла.
type PersistentStorage interface {
	SaveMetricToFile() error
	LoadMetricsFromFile() error
}

// GetterMetrics определяет интерфейс для получения метрик.
type GetterMetrics interface {
	GetGauges(ctx context.Context) map[string]float64
	GetCounters(ctx context.Context) map[string]int64
}

// Dump реализует сохранение и загрузку метрик в файл.
type Dump struct {
	storage GetterMetrics
	cfg     *config.Config
}

// MetricsDump — структура для сериализации метрик в файл.
type MetricsDump struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

// New создаёт новый объект Dump для работы с файлом метрик.
func New(storage GetterMetrics, cfg config.Config) *Dump {
	return &Dump{storage: storage, cfg: &cfg}
}

// SaveMetricToFile сохраняет метрики в файл с повторными попытками при ошибках.
func (d *Dump) SaveMetricToFile(ctx context.Context) error {
	dump := MetricsDump{
		Gauges:   d.storage.GetGauges(ctx),
		Counters: d.storage.GetCounters(ctx),
	}

	data, err := json.Marshal(dump)
	if err != nil {
		return err
	}

	for trying := 0; trying <= maxRetries; trying++ {
		err := os.WriteFile(d.cfg.FileStoragePath, data, 0644)
		if err != nil {
			if isFileLockedError(err) && trying < maxRetries {
				log.Printf("File is locked or unavailable for write (trying %d/%d): %v. Retrying in %v...", trying+1, maxRetries+1, err, retryDelays[trying])
				time.Sleep(retryDelays[trying])
				continue
			}
			log.Printf("error write file: %v", err)
			return err
		}
		return nil
	}

	return fmt.Errorf("error open file")
}

// LoadMetricsFromFile загружает метрики из файла.
func (d *Dump) LoadMetricsFromFile() error {
	var res []byte
	var err error
	for trying := 0; trying <= maxRetries; trying++ {
		res, err = os.ReadFile(d.cfg.FileStoragePath)
		if err != nil {
			if isFileLockedError(err) && trying < maxRetries {
				log.Printf("File is locked or unavailable for read (attempt %d/%d): %v. Retrying in %v...", trying+1, maxRetries+1, err, retryDelays[trying])
				time.Sleep(retryDelays[trying])
				continue
			}
			log.Printf("error open file: %v", err)
			return err
		}
	}

	err = json.Unmarshal(res, d.storage)
	if err != nil {
		return err
	}

	return nil
}

func isFileLockedError(err error) bool {
	if errors.Is(err, os.ErrPermission) {
		return true
	}

	if strings.Contains(err.Error(), "used by another process") ||
		strings.Contains(err.Error(), "resource temporarily unavailable") {
		return true
	}
	return false
}
