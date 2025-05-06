package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zubans/metrics/internal/logger"
	"go.uber.org/zap"
	"log"
	"time"
)

var DB *sql.DB

func InitDB(connStr string, migrationsPath string) error {
	const maxRetries = 3
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	if err := logger.Initialize("info"); err != nil {
		log.Printf("logger error: %v", err)
	}

	for trying := 0; trying < maxRetries; trying++ {
		m, err := migrate.New(
			fmt.Sprintf("file://%s", migrationsPath),
			connStr,
		)

		if err != nil {
			logger.Log.Info("Connection attempt failed",
				zap.Int("attempt", trying+1),
				zap.Error(err),
			)
			time.Sleep(getDelay(trying, retryDelays))
			continue
		}
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			if isConnectionError(err) {
				logger.Log.Info("Connection attempt failed",
					zap.Int("attempt", trying+1),
					zap.Error(err),
				)
				time.Sleep(getDelay(trying, retryDelays))
				continue
			}

			return fmt.Errorf("migrate.Up: %w", err)
		}

		DB, err = sql.Open("pgx", connStr)
		if err != nil {
			logger.Log.Info("Open sql failed",
				zap.Int("attempt", trying+1),
				zap.Error(err),
			)
			time.Sleep(getDelay(trying, retryDelays))
			continue
		}

	}

	return nil
}

func getDelay(try int, delays []time.Duration) time.Duration {
	if try >= len(delays) {
		return delays[len(delays)-1]
	}
	return delays[try]
}

func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.ConnectionException ||
			pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
			pgErr.Code == pgerrcode.ConnectionFailure
	}
	return false
}

func PingDB() error {
	return DB.Ping()
}
