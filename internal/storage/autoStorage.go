package storage

type AutoStorage struct {
	storage *MemStorage
	dump    *Dump
}

func NewAutoDump(storage *MemStorage, dump *Dump) *AutoStorage {
	return &AutoStorage{storage: storage, dump: dump}
}

func (s *AutoStorage) UpdateGauge(name string, value float64) float64 {
	res := s.storage.UpdateGauge(name, value)
	s.dump.SaveMetricToFile()

	return res
}

func (s *AutoStorage) UpdateCounter(name string, value int64) int64 {
	res := s.storage.UpdateCounter(name, value)
	s.dump.SaveMetricToFile()

	return res
}

func (s *AutoStorage) GetGauge(name string) (float64, bool) { return s.storage.GetGauge(name) }
func (s *AutoStorage) GetCounter(name string) (int64, bool) { return s.storage.GetCounter(name) }
func (s *AutoStorage) GetGauges() map[string]float64        { return s.storage.GetGauges() }
func (s *AutoStorage) GetCounters() map[string]int64        { return s.storage.GetCounters() }
func (s *AutoStorage) ShowMetrics() (map[string]float64, map[string]int64) {
	return s.storage.ShowMetrics()
}
