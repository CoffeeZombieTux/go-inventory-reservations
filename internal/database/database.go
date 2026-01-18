package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Database struct {
	DB     *sql.DB
	logger *logrus.Logger
}

func New(dsn string, logger *logrus.Logger) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
}

func (d *Database) Close() error {
	if d.DB != nil {
		d.logger.Info("Closing database connection")
		return d.DB.Close()
	}
	return nil
}

func (d *Database) Ping() error {
	return d.DB.Ping()
}

func (d *Database) HealthCheck() error {
	var result int
	err := d.DB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		d.logger.WithError(err).Error("Database health check failed")
		return err
	}
	return nil
}
