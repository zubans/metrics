package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/zubans/metrics/internal/models"
	"log"
	"time"
)

type PostDB struct {
	db *sql.DB
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
	_, err := db.db.ExecContext(ctx, "INSERT INTO metrics (type, name, delta, timestamp) VALUES ($1, $2, $3, $4) ON CONFLICT (name, type) DO UPDATE SET delta = $3", models.Counter, name, value, time.Now())

	if err != nil {
		log.Println("error insert metric: ", err)
	}

	return value
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

func (db *PostDB) GetGauges(ctx context.Context) map[string]float64 {
	var m models.MetricsDTO

	row := db.db.QueryRowContext(ctx, "select * from metrics where type = $1", models.Gauge)

	err := row.Scan(&m)
	if err != nil {
		log.Println("error get gauge metric: ", err)
		return nil
	}

	return nil
}

func (db *PostDB) GetCounters(ctx context.Context) map[string]int64 {
	var m models.MetricsDTO

	row := db.db.QueryRowContext(ctx, "select * from metrics where name = $1", models.Counter)

	err := row.Scan(&m)
	if err != nil {
		log.Println("error get gauge metric: ", err)
		return nil
	}

	return nil
}

func (db *PostDB) ShowMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	rows, err := db.db.QueryContext(ctx, "select name, type, value, delta from metrics")
	if err != nil {
		log.Println("Error querying metrics", err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			name        string
			metricType  string
			metricValue sql.NullFloat64
			delta       sql.NullInt64
		)

		if errors.Is(err, rows.Scan(&name, &metricType, &metricValue, &delta)); err != nil {
			log.Println("Error scanning row:", err)
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

	return gauges, counters
}
