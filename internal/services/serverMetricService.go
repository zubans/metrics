package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/zubans/metrics/internal/errdefs"
	"github.com/zubans/metrics/internal/models"
	"sort"
	"strconv"
)

type MetricStorage interface {
	UpdateGauge(name string, value float64) float64
	UpdateCounter(name string, value int64) int64
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetGauges() map[string]float64
	GetCounters() map[string]int64
	ShowMetrics() (map[string]float64, map[string]int64)
}

type Storage struct {
	storage MetricStorage
}

var validate = validator.New()

type MetricData struct {
	Type  string `validate:"required,oneof=counter gauge"`
	Name  string `validate:"required"`
	Value *string
}

func NewMetricData(t, n string, v ...string) (*MetricData, error) {
	m := &MetricData{
		Type: t,
		Name: n,
	}

	if len(v) > 0 {
		m.Value = &v[0]
	}

	if err := validate.Struct(m); err != nil {
		return nil, err
	}

	return m, nil
}

func ParseMetricValue(mData *MetricData) (float64, error) {
	if mData.Value == nil {
		return 0, fmt.Errorf("value is nil for metric %s", mData.Name)
	}

	value, err := strconv.ParseFloat(*mData.Value, 64)
	if err != nil {
		return 0, fmt.Errorf("can't parse Value '%s': %w", *mData.Value, err)
	}

	return value, nil
}

func (s Storage) ShowMetrics() string {
	gauges, counters := s.storage.ShowMetrics()

	var keys []string
	for k := range gauges {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := "<html><span style='float:left; margin-right:20px;'><strong>Guages</strong><table border=1 style='border-collapse: collapse;'><tr>"

	for _, k := range keys {
		v := gauges[k]
		result += "<tr><td>" + k + "</td><td>" + strconv.FormatFloat(v, 'f', -1, 64) + "</td></tr>"
	}

	result += "</table> </span>"
	result += "<span style='float:left;'><strong>Counters</strong><table border=1 style='border-collapse: collapse;'><tr>"

	for k, w := range counters {
		result += "<td>" + k + "</td>" + "<td>" + strconv.FormatInt(w, 10) + "</td></tr>"
	}
	result += "</table></span></html>"

	return result
}

func NewMetricService(storage MetricStorage) *Storage {
	return &Storage{storage: storage}
}

func (s Storage) GetMetric(mData *MetricData) (string, *errdefs.CustomError) {
	if mData.Type == "counter" {
		value, found := s.storage.GetCounter(mData.Name)
		if found {
			return strconv.FormatInt(value, 10), nil
		} else {
			return "", errdefs.NewNotFoundError("metric name required")
		}
	} else if mData.Type == "gauge" {
		value, found := s.storage.GetGauge(mData.Name)
		if found {
			return strconv.FormatFloat(value, 'f', -1, 64), nil
		} else {
			return "", errdefs.NewNotFoundError("metric name required")
		}
	} else {
		return "", errdefs.NewBadRequestError("Invalid metric type")
	}
}

func (s Storage) GetJSONMetric(jsonData *models.MetricsDTO) ([]byte, *errdefs.CustomError) {
	if jsonData.MType == string(models.Counter) {
		value, found := s.storage.GetCounter(jsonData.ID)
		if found {
			jsonData.Delta = &value
			res, err := json.Marshal(jsonData)
			if err != nil {
				return nil, errdefs.NewBadRequestError("can't marshal json data")
			}
			return res, nil
		} else {
			return nil, errdefs.NewNotFoundError("metric name required")
		}
	} else if jsonData.MType == string(models.Gauge) {
		value, found := s.storage.GetGauge(jsonData.ID)
		if found {
			jsonData.Value = &value
			res, err := json.Marshal(jsonData)
			if err != nil {
				return nil, errdefs.NewBadRequestError("can't marshal json data")
			}
			return res, nil
		} else {
			return nil, errdefs.NewNotFoundError("metric name required")
		}
	} else {
		return nil, errdefs.NewBadRequestError("Invalid metric type")
	}
}

func (s Storage) UpdateMetric(mData *MetricData) (*models.MetricsDTO, *errdefs.CustomError, error) {
	if mData.Name == "" {
		return nil, errdefs.NewNotFoundError("metric name required"), fmt.Errorf("metric name required")
	}

	switch mData.Type {
	case "gauge":
		value, err := ParseMetricValue(mData)
		if err != nil {
			return nil, errdefs.NewBadRequestError("invalid gauge value"), fmt.Errorf("invalid gauge value")
		}

		res := s.storage.UpdateGauge(mData.Name, value)

		return &models.MetricsDTO{
			ID:    mData.Name,
			MType: "gauge",
			Value: &res,
		}, nil, nil
	case "counter":
		value, err := ParseMetricValue(mData)
		if err != nil {
			return nil, errdefs.NewBadRequestError("invalid counter metric value"), fmt.Errorf("invalid counter metric value")
		}

		res := s.storage.UpdateCounter(mData.Name, int64(value))

		return &models.MetricsDTO{
			ID:    mData.Name,
			MType: "counter",
			Delta: &res,
		}, nil, nil
	default:
		return nil, errdefs.NewBadRequestError("invalid counter metric type"), fmt.Errorf("invalid counter metric type")
	}
}
