package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/zubans/metrics/internal/models"
)

type PostDB struct {
	db   *sql.DB
	dump *Dump
}

func NewDB(db *sql.DB) *PostDB {
	return &PostDB{db: db}
}

func (db *PostDB) UpdateGauge(ctx context.Context, name string, value float64) float64 {
	_, err := db.db.ExecContext(ctx, "INSERT INTO metrics (type, name, value, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (name, type) DO UPDATE SET value = $3", models.Gauge, name, value, time.Now())

	if err != nil {
		log.Println("error insert metric: ", err)
	}

	return value
}

func (db *PostDB) UpdateCounter(ctx context.Context, name string, value int64) int64 {
	_, err := db.db.ExecContext(ctx, "INSERT INTO metrics (type, name, delta, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (name, type) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, timestamp = EXCLUDED.timestamp", models.Counter, name, value, time.Now())

	if err != nil {
		log.Println("error insert metric: ", err)
	}

	return value
}

func (db *PostDB) UpdateMetrics(ctx context.Context, m []models.MetricsDTO) error {
	tx, err := db.db.Begin()
	if err != nil {
		log.Println("error create transaction:", err)
		return err
	}

	counterMap := make(map[string]int64)
	var gauges []models.MetricsDTO

	for _, v := range m {
		switch v.MType {
		case string(models.Counter):
			if v.Delta != nil {
				counterMap[v.ID] += *v.Delta
			}
		case string(models.Gauge):
			gauges = append(gauges, v)
		}
	}

	for k, v := range counterMap {
		_, err = tx.ExecContext(ctx, "INSERT INTO metrics (type, name, delta, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (name, type) DO UPDATE SET delta = $3", string(models.Counter), k, v, time.Now())
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}

	for _, v := range gauges {
		_, err = tx.ExecContext(ctx, "INSERT INTO metrics (type, name, value, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (name, type) DO UPDATE SET value = $3", string(models.Gauge), v.ID, v.Value, time.Now())
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *PostDB) GetGauge(ctx context.Context, name string) (float64, bool) {
	var m models.MetricsDTO

	row := db.db.QueryRowContext(ctx, "select name as id, type, value from metrics where name = $1 and type = $2 limit 1", name, models.Gauge)

	err := row.Scan(&m.ID, &m.MType, &m.Value)
	if err != nil {
		log.Println("error get gauge metric: ", err)
		return 0, false
	}

	return *m.Value, true
}

func (db *PostDB) GetCounter(ctx context.Context, name string) (int64, bool) {
	var m models.MetricsDTO

	row := db.db.QueryRowContext(ctx, "select name as id, type, delta from metrics where name = $1 and type = $2 limit 1", name, models.Counter)

	err := row.Scan(&m.ID, &m.MType, &m.Delta)
	if err != nil {
		log.Println("error get counter metric: ", err)
		return 0, false
	}

	return *m.Delta, true
}

func (db *PostDB) ShowMetrics(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	rows, err := db.db.QueryContext(ctx, "select name, type, value, delta from metrics")
	if err != nil {
		log.Println("Error querying metrics", err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println("Error rows close:", err)
		}
	}(rows)

	for rows.Next() {
		var (
			name        string
			metricType  string
			metricValue sql.NullFloat64
			delta       sql.NullInt64
		)

		err := rows.Scan(&name, &metricType, &metricValue, &delta)
		if err != nil {
			log.Printf("DATA LAYER: storage.postgres.GetAllMetrics: rows.Scan error: %v", err)
			continue
		}

		switch metricType {
		case string(models.Gauge):
			gauges[name] = metricValue.Float64
		case string(models.Counter):
			counters[name] = delta.Int64
		}
		if err = rows.Err(); err != nil {
			log.Println("Error after scanning rows", err)
		}
	}

	return gauges, counters, nil
}
