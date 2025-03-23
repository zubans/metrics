package storage

type MetricStorage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}
